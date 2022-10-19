package models

import "time"

type PicturesQueryParams struct {
	From      time.Time `form:"from" binding:"required,ltefield=To" time_format:"2006-01-02"`
	To        time.Time `form:"to" binding:"required" time_format:"2006-01-02"`
	Copyright *string   `form:"copyright"`
}
