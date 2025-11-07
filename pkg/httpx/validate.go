package httpx

import (
	"fmt"
	"reflect"
)

func SetDefaultPagination(query any) error {
	val := reflect.ValueOf(query)

	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("[httpx] : query must be a pointer")
	}

	val = val.Elem()

	paginationField := val.FieldByName("PaginationQuery")

	if paginationField.IsValid() {
		pagination := paginationField.Addr().Interface().(*PaginationQuery)

		if pagination.Page == nil {
			defaultPage := 1
			pagination.Page = &defaultPage
		} else if *pagination.Page <= 0 {
			*pagination.Page = 1
		}

		if pagination.PageSize == nil {
			defaultPageSize := 10
			pagination.PageSize = &defaultPageSize
		} else if *pagination.PageSize <= 0 {
			*pagination.PageSize = 10
		}
	}

	return nil
}
