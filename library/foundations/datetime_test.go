package foundations

import (
	"testing"
)

func TestNewDatetime(t *testing.T) {
	tests := []struct {
		name    string
		year    *int
		month   *int
		day     *int
		hour    *int
		minute  *int
		second  *int
		wantErr bool
	}{
		{
			name:  "full datetime",
			year:  intPtr(2024),
			month: intPtr(3),
			day:   intPtr(15),
			hour:  intPtr(10),
			minute: intPtr(30),
			second: intPtr(45),
		},
		{
			name:  "date only",
			year:  intPtr(2024),
			month: intPtr(12),
			day:   intPtr(25),
		},
		{
			name:   "time only",
			hour:   intPtr(14),
			minute: intPtr(30),
			second: intPtr(0),
		},
		{
			name:    "invalid month",
			year:    intPtr(2024),
			month:   intPtr(13),
			day:     intPtr(1),
			wantErr: true,
		},
		{
			name:    "invalid day for month",
			year:    intPtr(2024),
			month:   intPtr(2),
			day:     intPtr(30),
			wantErr: true,
		},
		{
			name:    "invalid hour",
			hour:    intPtr(25),
			minute:  intPtr(0),
			second:  intPtr(0),
			wantErr: true,
		},
		{
			name:  "leap year feb 29",
			year:  intPtr(2024),
			month: intPtr(2),
			day:   intPtr(29),
		},
		{
			name:    "non-leap year feb 29",
			year:    intPtr(2023),
			month:   intPtr(2),
			day:     intPtr(29),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, err := NewDatetime(tt.year, tt.month, tt.day, tt.hour, tt.minute, tt.second)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatetime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && dt == nil {
				t.Error("NewDatetime() returned nil without error")
			}
		})
	}
}

func TestDatetimeAccessors(t *testing.T) {
	dt, err := NewFullDatetime(2024, 3, 15, 10, 30, 45)
	if err != nil {
		t.Fatalf("NewFullDatetime() error = %v", err)
	}

	if got := dt.YearOr(0); got != 2024 {
		t.Errorf("Year() = %v, want 2024", got)
	}
	if got := dt.MonthOr(0); got != 3 {
		t.Errorf("Month() = %v, want 3", got)
	}
	if got := dt.DayOr(0); got != 15 {
		t.Errorf("Day() = %v, want 15", got)
	}
	if got := dt.HourOr(0); got != 10 {
		t.Errorf("Hour() = %v, want 10", got)
	}
	if got := dt.MinuteOr(0); got != 30 {
		t.Errorf("Minute() = %v, want 30", got)
	}
	if got := dt.SecondOr(0); got != 45 {
		t.Errorf("Second() = %v, want 45", got)
	}
}

func TestDatetimeWeekday(t *testing.T) {
	tests := []struct {
		name string
		year int
		month int
		day  int
		want int
	}{
		{"Monday", 2024, 1, 1, 1},      // January 1, 2024 is Monday
		{"Tuesday", 2024, 1, 2, 2},
		{"Wednesday", 2024, 1, 3, 3},
		{"Thursday", 2024, 1, 4, 4},
		{"Friday", 2024, 1, 5, 5},
		{"Saturday", 2024, 1, 6, 6},
		{"Sunday", 2024, 1, 7, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, _ := NewDate(tt.year, tt.month, tt.day)
			wd := dt.Weekday()
			if wd == nil {
				t.Error("Weekday() returned nil")
				return
			}
			if *wd != tt.want {
				t.Errorf("Weekday() = %v, want %v", *wd, tt.want)
			}
		})
	}
}

func TestDatetimeOrdinal(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month int
		day   int
		want  int
	}{
		{"Jan 1", 2024, 1, 1, 1},
		{"Jan 31", 2024, 1, 31, 31},
		{"Feb 1", 2024, 2, 1, 32},
		{"Mar 1 leap year", 2024, 3, 1, 61},
		{"Dec 31 leap year", 2024, 12, 31, 366},
		{"Dec 31 non-leap", 2023, 12, 31, 365},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, _ := NewDate(tt.year, tt.month, tt.day)
			ord := dt.Ordinal()
			if ord == nil {
				t.Error("Ordinal() returned nil")
				return
			}
			if *ord != tt.want {
				t.Errorf("Ordinal() = %v, want %v", *ord, tt.want)
			}
		})
	}
}

func TestDatetimeDisplay(t *testing.T) {
	tests := []struct {
		name    string
		dt      *Datetime
		pattern string
		want    string
	}{
		{
			name:    "full datetime default",
			dt:      mustDatetime(NewFullDatetime(2024, 3, 15, 10, 30, 45)),
			pattern: "",
			want:    "2024-03-15 10:30:45",
		},
		{
			name:    "date only default",
			dt:      mustDatetime(NewDate(2024, 3, 15)),
			pattern: "",
			want:    "2024-03-15",
		},
		{
			name:    "time only default",
			dt:      mustDatetime(NewTime(10, 30, 45)),
			pattern: "",
			want:    "10:30:45",
		},
		{
			name:    "custom pattern",
			dt:      mustDatetime(NewFullDatetime(2024, 3, 15, 10, 30, 45)),
			pattern: "[day]/[month]/[year]",
			want:    "15/03/2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dt.Format(tt.pattern)
			if got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDuration(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		minutes int64
		hours   int64
		days    int64
		weeks   int64
		wantSec float64
	}{
		{"1 second", 1, 0, 0, 0, 0, 1},
		{"1 minute", 0, 1, 0, 0, 0, 60},
		{"1 hour", 0, 0, 1, 0, 0, 3600},
		{"1 day", 0, 0, 0, 1, 0, 86400},
		{"1 week", 0, 0, 0, 0, 1, 604800},
		{"combined", 30, 5, 2, 1, 0, 30 + 300 + 7200 + 86400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDuration(tt.seconds, tt.minutes, tt.hours, tt.days, tt.weeks)
			if got := d.Seconds(); got != tt.wantSec {
				t.Errorf("Seconds() = %v, want %v", got, tt.wantSec)
			}
		})
	}
}

func TestDurationAccessors(t *testing.T) {
	// 1 week = 604800 seconds
	d := NewDuration(0, 0, 0, 0, 1)

	if got := d.Weeks(); got != 1 {
		t.Errorf("Weeks() = %v, want 1", got)
	}
	if got := d.Days(); got != 7 {
		t.Errorf("Days() = %v, want 7", got)
	}
	if got := d.Hours(); got != 168 {
		t.Errorf("Hours() = %v, want 168", got)
	}
	if got := d.Minutes(); got != 10080 {
		t.Errorf("Minutes() = %v, want 10080", got)
	}
	if got := d.Seconds(); got != 604800 {
		t.Errorf("Seconds() = %v, want 604800", got)
	}
}

func TestDatetimeArithmetic(t *testing.T) {
	dt1, _ := NewFullDatetime(2024, 3, 15, 10, 30, 0)
	dt2, _ := NewFullDatetime(2024, 3, 15, 11, 30, 0)
	oneHour := NewDuration(0, 0, 1, 0, 0)

	// datetime + duration
	result, err := Add(dt1, oneHour)
	if err != nil {
		t.Fatalf("Add(datetime, duration) error = %v", err)
	}
	resultDt, ok := result.(*Datetime)
	if !ok {
		t.Fatal("Add(datetime, duration) did not return a datetime")
	}
	if resultDt.HourOr(0) != 11 {
		t.Errorf("Add(datetime, duration) hour = %v, want 11", resultDt.HourOr(0))
	}

	// duration + datetime (commutative)
	result, err = Add(oneHour, dt1)
	if err != nil {
		t.Fatalf("Add(duration, datetime) error = %v", err)
	}
	resultDt, ok = result.(*Datetime)
	if !ok {
		t.Fatal("Add(duration, datetime) did not return a datetime")
	}
	if resultDt.HourOr(0) != 11 {
		t.Errorf("Add(duration, datetime) hour = %v, want 11", resultDt.HourOr(0))
	}

	// datetime - duration
	result, err = Sub(dt2, oneHour)
	if err != nil {
		t.Fatalf("Sub(datetime, duration) error = %v", err)
	}
	resultDt, ok = result.(*Datetime)
	if !ok {
		t.Fatal("Sub(datetime, duration) did not return a datetime")
	}
	if resultDt.HourOr(0) != 10 {
		t.Errorf("Sub(datetime, duration) hour = %v, want 10", resultDt.HourOr(0))
	}

	// datetime - datetime = duration
	result, err = Sub(dt2, dt1)
	if err != nil {
		t.Fatalf("Sub(datetime, datetime) error = %v", err)
	}
	resultDur, ok := result.(Duration)
	if !ok {
		t.Fatal("Sub(datetime, datetime) did not return a duration")
	}
	if resultDur.Hours() != 1 {
		t.Errorf("Sub(datetime, datetime) = %v hours, want 1", resultDur.Hours())
	}
}

func TestDurationArithmetic(t *testing.T) {
	d1 := NewDuration(30, 0, 0, 0, 0)
	d2 := NewDuration(15, 0, 0, 0, 0)

	// duration + duration
	result, err := Add(d1, d2)
	if err != nil {
		t.Fatalf("Add(duration, duration) error = %v", err)
	}
	resultDur, ok := result.(Duration)
	if !ok {
		t.Fatal("Add(duration, duration) did not return a duration")
	}
	if resultDur.Seconds() != 45 {
		t.Errorf("Add(duration, duration) = %v seconds, want 45", resultDur.Seconds())
	}

	// duration - duration
	result, err = Sub(d1, d2)
	if err != nil {
		t.Fatalf("Sub(duration, duration) error = %v", err)
	}
	resultDur, ok = result.(Duration)
	if !ok {
		t.Fatal("Sub(duration, duration) did not return a duration")
	}
	if resultDur.Seconds() != 15 {
		t.Errorf("Sub(duration, duration) = %v seconds, want 15", resultDur.Seconds())
	}
}

func TestDatetimeEquality(t *testing.T) {
	dt1, _ := NewFullDatetime(2024, 3, 15, 10, 30, 0)
	dt2, _ := NewFullDatetime(2024, 3, 15, 10, 30, 0)
	dt3, _ := NewFullDatetime(2024, 3, 15, 10, 30, 1)

	// Equal datetimes
	result, _ := Eq(dt1, dt2)
	if result != Bool(true) {
		t.Error("Eq(dt1, dt2) should be true")
	}

	// Unequal datetimes
	result, _ = Eq(dt1, dt3)
	if result != Bool(false) {
		t.Error("Eq(dt1, dt3) should be false")
	}
}

func TestDatetimeComparison(t *testing.T) {
	dt1, _ := NewFullDatetime(2024, 3, 15, 10, 30, 0)
	dt2, _ := NewFullDatetime(2024, 3, 15, 11, 30, 0)

	// Lt
	result, err := Lt(dt1, dt2)
	if err != nil {
		t.Fatalf("Lt() error = %v", err)
	}
	if result != Bool(true) {
		t.Error("Lt(dt1, dt2) should be true")
	}

	// Gt
	result, err = Gt(dt2, dt1)
	if err != nil {
		t.Fatalf("Gt() error = %v", err)
	}
	if result != Bool(true) {
		t.Error("Gt(dt2, dt1) should be true")
	}
}

func TestDurationEquality(t *testing.T) {
	d1 := NewDuration(60, 0, 0, 0, 0)
	d2 := NewDuration(0, 1, 0, 0, 0) // 1 minute = 60 seconds

	result, _ := Eq(d1, d2)
	if result != Bool(true) {
		t.Error("Eq(60s, 1min) should be true")
	}
}

func TestDurationComparison(t *testing.T) {
	d1 := NewDuration(30, 0, 0, 0, 0)
	d2 := NewDuration(60, 0, 0, 0, 0)

	result, err := Lt(d1, d2)
	if err != nil {
		t.Fatalf("Lt() error = %v", err)
	}
	if result != Bool(true) {
		t.Error("Lt(30s, 60s) should be true")
	}
}

func TestToday(t *testing.T) {
	dt := Today(nil)
	if dt == nil {
		t.Fatal("Today() returned nil")
	}
	if !dt.HasDate() {
		t.Error("Today() should have date components")
	}
	if dt.HasTime() {
		t.Error("Today() should not have time components")
	}
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func mustDatetime(dt *Datetime, err error) *Datetime {
	if err != nil {
		panic(err)
	}
	return dt
}
