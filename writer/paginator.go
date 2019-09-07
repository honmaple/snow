/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: paginator.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-04-04 15:41:27 (CST)
 Last Update: Saturday 2019-09-07 02:15:26 (CST)
		  By:
 Description:
 *********************************************************************************/
package writer


// Paginator ..
type Paginator struct {
	Items  []map[string]interface{}
	Number int
}

// List ..
func (s *Paginator) List() []map[string]interface{} {
	pages := s.Pages()
	length := len(s.Items)
	paginator := make([]map[string]interface{}, pages)
	for i := 0; i < pages; i++ {
		page := i + 1
		end := page * s.Number
		if end > length {
			end = length
		}
		paginator[i] = map[string]interface{}{
			"has_next":        page < pages,
			"has_previous":    page < pages && page > 1,
			"has_other_pages": page < pages || (page < pages && page > 1),
			"next_page":       page < pages || (page < pages && page > 1),
			"previous_page":   page < pages || (page < pages && page > 1),
			"page":            page,
			"pages":           pages,
			"object_list":     s.Items[(page-1)*s.Number : end],
		}
	}
	return paginator
}

// Pages ..
func (s *Paginator) Pages() int {
	length := len(s.Items)
	if length%s.Number == 0 {
		return length / s.Number
	}
	return length/s.Number + 1
}
