package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

// NextDate вычисляет следующую дату выполнения задачи
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	// Строгая валидация даты начала
	startDate, err := time.Parse(DateFormat, dstart)
	if err != nil || startDate.Format(DateFormat) != dstart {
		return "", fmt.Errorf("некорректная дата начала: %s", dstart)
	}

	// Обнуляем время для корректного сравнения только по дням
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	date := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)

	parts := strings.SplitN(repeat, " ", 2)
	rule := parts[0]

	switch rule {
	case "d":
		if len(parts) < 2 {
			return "", fmt.Errorf("для правила 'd' не указан интервал")
		}
		interval, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || interval < 1 || interval > 400 {
			return "", fmt.Errorf("неверный интервал для 'd' (допустимо 1-400)")
		}
		for {
			date = date.AddDate(0, 0, interval)
			if date.After(now) {
				return date.Format(DateFormat), nil
			}
		}

	case "y":
		for {
			date = date.AddDate(1, 0, 0)
			if date.After(now) {
				return date.Format(DateFormat), nil
			}
		}

	case "w":
		if len(parts) < 2 {
			return "", fmt.Errorf("для правила 'w' не указаны дни недели")
		}
		days := make([]bool, 8) // индексы 1-7
		for _, s := range strings.Split(parts[1], ",") {
			d, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || d < 1 || d > 7 {
				return "", fmt.Errorf("недопустимый день недели: %s", s)
			}
			days[d] = true
		}
		for {
			wd := int(date.Weekday())
			if wd == 0 {
				wd = 7 // Воскресенье -> 7
			}
			if date.After(now) && days[wd] {
				return date.Format(DateFormat), nil
			}
			date = date.AddDate(0, 0, 1)
		}

	case "m":
		if len(parts) < 2 {
			return "", fmt.Errorf("для правила 'm' не указаны дни месяца")
		}
		mParts := strings.SplitN(parts[1], " ", 2)

		days := make([]bool, 34) // 1-31, 32=последний, 33=предпоследний
		for _, s := range strings.Split(mParts[0], ",") {
			d, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				return "", fmt.Errorf("неверный день месяца: %s", s)
			}
			if d >= 1 && d <= 31 {
				days[d] = true
			} else if d == -1 {
				days[32] = true
			} else if d == -2 {
				days[33] = true
			} else {
				return "", fmt.Errorf("день месяца должен быть 1-31, -1 или -2")
			}
		}

		months := make([]bool, 13)
		if len(mParts) > 1 {
			for _, s := range strings.Split(mParts[1], ",") {
				m, err := strconv.Atoi(strings.TrimSpace(s))
				if err != nil || m < 1 || m > 12 {
					return "", fmt.Errorf("недопустимый месяц: %s", s)
				}
				months[m] = true
			}
		} else {
			for i := 1; i <= 12; i++ {
				months[i] = true
			}
		}

		for {
			day := date.Day()
			month := int(date.Month())
			// Последний день текущего месяца
			eom := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			isLast := day == eom
			isSecLast := day == eom-1

			matchDay := days[day] || (isLast && days[32]) || (isSecLast && days[33])
			matchMonth := months[month]

			if date.After(now) && matchDay && matchMonth {
				return date.Format(DateFormat), nil
			}
			date = date.AddDate(0, 0, 1)
		}

	default:
		return "", fmt.Errorf("неподдерживаемый формат правила: %s", rule)
	}
}

// NextDateHandler обрабатывает GET 
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			http.Error(w, "Неверный формат параметра now", http.StatusBadRequest)
			return
		}
	}

	result, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, result)
}