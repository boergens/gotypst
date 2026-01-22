package foundations

import (
	"fmt"
	"strings"
	"time"
)

// NewDatetime creates a new datetime with optional components.
// Pass nil for any component that should be unset.
func NewDatetime(year, month, day, hour, minute, second *int) (*Datetime, error) {
	dt := &Datetime{
		year:   year,
		month:  month,
		day:    day,
		hour:   hour,
		minute: minute,
		second: second,
	}

	// Validate the datetime
	if err := dt.validate(); err != nil {
		return nil, err
	}

	return dt, nil
}

// NewDate creates a date-only datetime.
func NewDate(year, month, day int) (*Datetime, error) {
	y, m, d := year, month, day
	return NewDatetime(&y, &m, &d, nil, nil, nil)
}

// NewTime creates a time-only datetime.
func NewTime(hour, minute, second int) (*Datetime, error) {
	h, mi, s := hour, minute, second
	return NewDatetime(nil, nil, nil, &h, &mi, &s)
}

// NewFullDatetime creates a datetime with all components.
func NewFullDatetime(year, month, day, hour, minute, second int) (*Datetime, error) {
	y, mo, d := year, month, day
	h, mi, s := hour, minute, second
	return NewDatetime(&y, &mo, &d, &h, &mi, &s)
}

// Today returns the current date.
// The offset parameter specifies the UTC offset in hours (nil for local time).
func Today(offset *int) *Datetime {
	var t time.Time
	if offset != nil {
		loc := time.FixedZone("", *offset*3600)
		t = time.Now().In(loc)
	} else {
		t = time.Now()
	}

	y, m, d := t.Year(), int(t.Month()), t.Day()
	return &Datetime{
		year:  &y,
		month: &m,
		day:   &d,
	}
}

// validate checks if the datetime components form a valid date.
func (dt *Datetime) validate() error {
	if dt == nil {
		return nil
	}

	// Validate month
	if dt.month != nil {
		m := *dt.month
		if m < 1 || m > 12 {
			return &OpError{Message: fmt.Sprintf("month must be between 1 and 12, got %d", m)}
		}
	}

	// Validate day
	if dt.day != nil {
		d := *dt.day
		if d < 1 {
			return &OpError{Message: fmt.Sprintf("day must be at least 1, got %d", d)}
		}
		// Check against days in month if month is specified
		if dt.month != nil && dt.year != nil {
			maxDay := daysInMonth(*dt.year, *dt.month)
			if d > maxDay {
				return &OpError{Message: fmt.Sprintf("day %d is invalid for month %d", d, *dt.month)}
			}
		} else if d > 31 {
			return &OpError{Message: fmt.Sprintf("day must be at most 31, got %d", d)}
		}
	}

	// Validate time components
	if dt.hour != nil {
		h := *dt.hour
		if h < 0 || h > 23 {
			return &OpError{Message: fmt.Sprintf("hour must be between 0 and 23, got %d", h)}
		}
	}
	if dt.minute != nil {
		m := *dt.minute
		if m < 0 || m > 59 {
			return &OpError{Message: fmt.Sprintf("minute must be between 0 and 59, got %d", m)}
		}
	}
	if dt.second != nil {
		s := *dt.second
		if s < 0 || s > 59 {
			return &OpError{Message: fmt.Sprintf("second must be between 0 and 59, got %d", s)}
		}
	}

	return nil
}

// daysInMonth returns the number of days in the given month.
func daysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if isLeapYear(year) {
			return 29
		}
		return 28
	default:
		return 0
	}
}

// isLeapYear checks if the year is a leap year.
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// Weekday returns the day of the week (Monday=1, Sunday=7) or nil if
// the datetime doesn't have complete date information.
func (dt *Datetime) Weekday() *int {
	if dt == nil || dt.year == nil || dt.month == nil || dt.day == nil {
		return nil
	}

	// Use Go's time package to compute weekday
	t := time.Date(*dt.year, time.Month(*dt.month), *dt.day, 0, 0, 0, 0, time.UTC)
	wd := int(t.Weekday())
	// Convert from Go's Sunday=0 to Typst's Monday=1
	if wd == 0 {
		wd = 7
	}
	return &wd
}

// Ordinal returns the day of the year (1-366) or nil if
// the datetime doesn't have complete date information.
func (dt *Datetime) Ordinal() *int {
	if dt == nil || dt.year == nil || dt.month == nil || dt.day == nil {
		return nil
	}

	t := time.Date(*dt.year, time.Month(*dt.month), *dt.day, 0, 0, 0, 0, time.UTC)
	ord := t.YearDay()
	return &ord
}

// Format formats the datetime using the given pattern.
// If pattern is empty, uses a default format based on which components are present.
func (dt *Datetime) Format(pattern string) string {
	if dt == nil {
		return ""
	}

	hasDate := dt.HasDate()
	hasTime := dt.HasTime()

	// Use default pattern if not specified
	if pattern == "" {
		if hasDate && hasTime {
			pattern = "[year]-[month]-[day] [hour]:[minute]:[second]"
		} else if hasDate {
			pattern = "[year]-[month]-[day]"
		} else if hasTime {
			pattern = "[hour]:[minute]:[second]"
		} else {
			return ""
		}
	}

	// Simple pattern replacement
	result := pattern

	if dt.year != nil {
		result = strings.ReplaceAll(result, "[year]", fmt.Sprintf("%04d", *dt.year))
	}
	if dt.month != nil {
		result = strings.ReplaceAll(result, "[month]", fmt.Sprintf("%02d", *dt.month))
	}
	if dt.day != nil {
		result = strings.ReplaceAll(result, "[day]", fmt.Sprintf("%02d", *dt.day))
	}
	if dt.hour != nil {
		result = strings.ReplaceAll(result, "[hour]", fmt.Sprintf("%02d", *dt.hour))
	}
	if dt.minute != nil {
		result = strings.ReplaceAll(result, "[minute]", fmt.Sprintf("%02d", *dt.minute))
	}
	if dt.second != nil {
		result = strings.ReplaceAll(result, "[second]", fmt.Sprintf("%02d", *dt.second))
	}

	// Handle weekday
	if wd := dt.Weekday(); wd != nil {
		weekdays := []string{"", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
		result = strings.ReplaceAll(result, "[weekday]", weekdays[*wd])
	}

	return result
}

// ToTime converts the datetime to a Go time.Time.
// Missing components default to sensible values (year 0, month 1, day 1, 00:00:00).
func (dt *Datetime) ToTime() time.Time {
	if dt == nil {
		return time.Time{}
	}

	year := dt.YearOr(0)
	month := dt.MonthOr(1)
	day := dt.DayOr(1)
	hour := dt.HourOr(0)
	minute := dt.MinuteOr(0)
	second := dt.SecondOr(0)

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
}

// NewDuration creates a duration from component values.
// The total duration is the sum of all components.
func NewDuration(seconds, minutes, hours, days, weeks int64) Duration {
	total := seconds * 1e9
	total += minutes * 60 * 1e9
	total += hours * 3600 * 1e9
	total += days * 86400 * 1e9
	total += weeks * 604800 * 1e9
	return Duration(total)
}

// DurationFromSeconds creates a duration from a number of seconds.
func DurationFromSeconds(s int64) Duration {
	return Duration(s * 1e9)
}

// DurationFromNanoseconds creates a duration from nanoseconds.
func DurationFromNanoseconds(ns int64) Duration {
	return Duration(ns)
}
