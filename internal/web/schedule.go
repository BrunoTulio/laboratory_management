package web

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"laboratory-management/internal/booking"
)

// SlotView represents a single time slot in the UI (free or booked).
type SlotView struct {
	Number      int
	LessonLabel string
	Label       string
	Start       string
	End         string
	Booked      bool
	TeacherName string
	ClassName   string
	SubjectName string
	Reason      string
}

// ShiftSummary represents one shift column/summary (Manhã, Tarde, Noite).
type ShiftSummary struct {
	Key          string
	Title        string
	BookedCount  int
	FreeCount    int
	FreeLabels   string
	Slots        []SlotView
	ShowInFilter bool
}

// WeekDaySchedule holds the daily summary for one day in the week view.
type WeekDaySchedule struct {
	Date          time.Time
	DateISO       string
	DateFormatted string
	Summaries     []ShiftSummary
	BookedCount   int
	FreeCount     int
}

var defaultShiftDefs = []struct {
	Key   string
	Title string
	Slots [][2]string
}{
	{
		Key:   "MATUTINO",
		Title: "Manhã · 07h30–11h30",
		Slots: [][2]string{
			{"07:30", "08:15"},
			{"08:15", "09:00"},
			{"09:15", "10:00"},
			{"10:00", "10:45"},
			{"10:45", "11:30"},
		},
	},
	{
		Key:   "VESPERTINO",
		Title: "Tarde · 13h–17h",
		Slots: [][2]string{
			{"13:00", "13:45"},
			{"13:45", "14:30"},
			{"14:30", "15:15"},
			{"15:30", "16:15"},
			{"16:15", "17:00"},
		},
	},
	{
		Key:   "NOTURNO",
		Title: "Noite · 18h50–21h50",
		Slots: [][2]string{
			{"18:50", "19:20"},
			{"19:20", "19:50"},
			{"19:50", "20:20"},
			{"20:20", "20:50"},
			{"20:50", "21:20"},
			{"21:20", "21:50"},
		},
	},
}

// BuildShiftSummaries computes shift slot availability and booking counts from a list of bookings.
func BuildShiftSummaries(bookings []*booking.Booking, filterShift, filterTeacher, filterClassID, filterSubject string) ([]ShiftSummary, int, int) {
	filtered := filterBookings(bookings, filterTeacher, filterClassID, filterSubject)

	var summaries []ShiftSummary
	totalFree := 0
	totalBooked := 0

	for _, def := range defaultShiftDefs {
		summary := ShiftSummary{
			Key:          def.Key,
			Title:        def.Title,
			ShowInFilter: filterShift == "" || filterShift == def.Key,
		}

		var freeLabels []string
		for idx, pair := range def.Slots {
			slotStart, _ := time.Parse("15:04", pair[0])
			slotEnd, _ := time.Parse("15:04", pair[1])
			num := idx + 1

			slot := SlotView{
				Number:      num,
				LessonLabel: fmt.Sprintf("%dª aula", num),
				Label:       fmt.Sprintf("%s–%s", pair[0], pair[1]),
				Start:       pair[0],
				End:         pair[1],
			}

			for _, b := range filtered {
				bStart := normalizeClock(b.StartTime)
				bEnd := normalizeClock(b.EndTime)
				if bStart.Before(slotEnd) && bEnd.After(slotStart) {
					slot.Booked = true
					slot.TeacherName = b.TeacherName
					slot.ClassName = b.ClassName
					slot.SubjectName = b.SubjectName
					slot.Reason = b.Reason
					break
				}
			}

			if slot.Booked {
				summary.BookedCount++
				totalBooked++
			} else {
				summary.FreeCount++
				totalFree++
				freeLabels = append(freeLabels, slot.Label)
			}

			summary.Slots = append(summary.Slots, slot)
		}

		summary.FreeLabels = strings.Join(freeLabels, " · ")
		summaries = append(summaries, summary)
	}

	return summaries, totalBooked, totalFree
}

func filterBookings(bookings []*booking.Booking, teacher, classFilter, subject string) []*booking.Booking {
	var out []*booking.Booking
	for _, b := range bookings {
		if teacher != "" && !strings.EqualFold(b.TeacherName, teacher) {
			continue
		}
		if classFilter != "" {
			matchID := b.ClassID != nil && fmt.Sprintf("%d", *b.ClassID) == classFilter
			matchName := strings.EqualFold(b.ClassName, classFilter)
			if !matchID && !matchName {
				continue
			}
		}
		if subject != "" && !strings.EqualFold(b.SubjectName, subject) {
			continue
		}
		out = append(out, b)
	}
	return out
}

func normalizeClock(t time.Time) time.Time {
	return time.Date(0, 1, 1, t.Hour(), t.Minute(), 0, 0, time.UTC)
}

// BuildWeekDates returns Monday through Saturday of the week containing target.
func BuildWeekDates(target time.Time) []time.Time {
	weekday := int(target.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := target.AddDate(0, 0, -(weekday - 1))
	days := make([]time.Time, 0, 6)
	for i := 0; i < 6; i++ {
		days = append(days, monday.AddDate(0, 0, i))
	}
	return days
}

// FormatDatePTBR formats a date in Brazilian Portuguese short format (e.g., "ter., 17 de nov. de 2026").
func FormatDatePTBR(t time.Time) string {
	weekdays := []string{"dom.", "seg.", "ter.", "qua.", "qui.", "sex.", "sáb."}
	months := []string{"jan.", "fev.", "mar.", "abr.", "mai.", "jun.", "jul.", "ago.", "set.", "out.", "nov.", "dez."}
	return fmt.Sprintf("%s, %d de %s de %d", weekdays[t.Weekday()], t.Day(), months[t.Month()-1], t.Year())
}

// ResolveLessonRange converts a shift ("MATUTINO", "VESPERTINO", "NOTURNO") and selected lesson numbers (1..6)
// into start and end times covering the selected lessons.
func ResolveLessonRange(shift string, lessonParams []string) (time.Time, time.Time, error) {
	if len(lessonParams) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("selecione pelo menos uma aula")
	}

	var slots [][2]string
	for _, def := range defaultShiftDefs {
		if def.Key == shift {
			slots = def.Slots
			break
		}
	}
	if len(slots) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("período inválido")
	}

	var nums []int
	for _, raw := range lessonParams {
		n, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || n < 1 || n > len(slots) {
			return time.Time{}, time.Time{}, fmt.Errorf("aula inválida para o período selecionado")
		}
		nums = append(nums, n)
	}
	sort.Ints(nums)

	startTime, err := time.Parse("15:04", slots[nums[0]-1][0])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endTime, err := time.Parse("15:04", slots[nums[len(nums)-1]-1][1])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return startTime, endTime, nil
}
