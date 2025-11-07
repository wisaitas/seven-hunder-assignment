package httpx

import (
	"context"

	"gorm.io/gorm"
)

func ProbeNextPages(
	ctx context.Context,
	column string,
	probeQuery *gorm.DB,
	orderBy string,
	page, pageSize, wantPages int,
) (int, error) {
	if wantPages <= 0 {
		return 0, nil
	}

	// ต้องการเช็คได้ไกลสุดกี่ "หน้า" ทางขวา
	need := wantPages*pageSize + 1

	// สร้าง struct สำหรับรับค่า id เท่านั้น
	type IDResult struct {
		ID string `gorm:"column:id"`
	}

	var results []IDResult

	// สร้าง query ใหม่โดยเลือกเฉพาะ column ที่ต้องการ
	// และใช้ order by, offset, limit
	query := probeQuery.WithContext(ctx).
		Select(column + " as id").
		Order(orderBy).
		Offset((page - 1) * pageSize).
		Limit(need)

	if err := query.Scan(&results).Error; err != nil {
		return 0, err
	}

	// จำนวนหน้าที่ไปได้จริงทางขวา
	remain := len(results)

	// ต้องหักหน้าปัจจุบันออกก่อน (pageSize แรก คือหน้าปัจจุบัน)
	if remain > pageSize {
		remain = remain - pageSize
	} else {
		// ไม่มีข้อมูลเกินหน้าปัจจุบัน
		return 0, nil
	}

	availPages := remain / pageSize
	if availPages > wantPages {
		availPages = wantPages
	}
	if availPages < 0 {
		availPages = 0
	}
	return availPages, nil
}

// pageWindowClamped: ตรึงขอบด้วย hasPrev/hasNext + nextPagesAvail (จำนวน "หน้า" ที่ยังไปทางขวาได้จริง)
func PageWindowClamped(current, width int, hasPrev, hasNext bool, nextPagesAvail int) []int {
	if width <= 0 {
		return []int{}
	}
	if current < 1 {
		current = 1
	}
	half := width / 2

	// 1) ถ้าชนซ้าย (หน้าแรก) ให้เริ่มที่ 1 แล้ว "end" = 1 + min(width-1, nextPagesAvail)
	if !hasPrev {
		span := width - 1
		if nextPagesAvail < span {
			span = nextPagesAvail
		}
		start := 1
		end := start + span
		out := make([]int, 0, end-start+1)
		for i := start; i <= end; i++ {
			out = append(out, i)
		}
		return out
	}

	// 2) ถ้าชนขวา ให้จบที่ current แล้วไล่ซ้ายให้ครบความกว้าง
	if !hasNext {
		end := current
		start := end - (width - 1)
		if start < 1 {
			start = 1
		}
		out := make([]int, 0, end-start+1)
		for i := start; i <= end; i++ {
			out = append(out, i)
		}
		return out
	}

	// 3) อยู่กลาง: ปกติขยับขวาได้ครึ่งหน้าต่าง แต่ clamp ด้วย nextPagesAvail
	right := half
	if nextPagesAvail < right {
		right = nextPagesAvail
	}
	end := current + right
	start := end - (width - 1)
	if start < 1 {
		start = 1
		end = start + (width - 1)
	}

	out := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		out = append(out, i)
	}
	return out
}
