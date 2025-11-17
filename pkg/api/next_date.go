package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ljnsur/todosay/pkg/constants"
	applog "github.com/ljnsur/todosay/pkg/log"
)

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	applog.Printf("nextDay: начало обработки %s %s", r.Method, r.RemoteAddr)
	if r.Method != http.MethodGet {
		applog.Printf("nextDay: неверный метод %s", r.Method)
		http.Error(w, "разрешены только GET запросы ", http.StatusMethodNotAllowed)
		return
	}

	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")
	if now == "" {
		now = time.Now().String()
	}
	nowTime, err := time.Parse(constants.TimeFormat, now)
	if err != nil {
		applog.Printf("nextDay: неверный формат now: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(nowTime, date, repeat)
	if err != nil {
		applog.Printf("nextDay: ошибка вычисления nextDate: %v", err)
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	nextDateN, err := strconv.Atoi(nextDate)
	if err != nil {
		applog.Printf("nextDay: преобразование nextDate в int не удалось: %v", err)
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	applog.Printf("nextDay: возвращаю nextDate=%d", nextDateN)
	writeJson(w, http.StatusOK, nextDateN)

}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {

	applog.Printf("NextDate: вычисление следующей даты now=%s dstart=%s repeat=%s", now.Format(constants.TimeFormat), dstart, repeat)
	dstartTime, err := time.Parse(constants.TimeFormat, dstart)
	if err != nil {
		applog.Printf("NextDate: ошибка парсинга dstart: %v", err)
		return "", err
	}

	if repeat == "" {
		applog.Printf("NextDate: правило повтора пустое")
		return "", errors.New("отсутсвует правило повторения ")
	}

	rule, err := parseRepeatRule(repeat)
	if err != nil {
		applog.Printf("NextDate: неверное правило повтора %q: %v", repeat, err)
		return "", err
	}

	nextDate := dstartTime

	switch rule.Type {
	case "d":
		for {
			nextDate = nextDate.AddDate(0, 0, rule.Days[0])
			if afterNow(nextDate, now) {
				applog.Printf("NextDate: найдено nextDate=%s для типа d", nextDate.Format(constants.TimeFormat))
				return nextDate.Format(constants.TimeFormat), nil
			}
		}
	case "w":
		for {
			nextDate = nextDate.AddDate(0, 0, 1)
			if afterNow(nextDate, now) {
				day := ruWeekDay(nextDate.Weekday())
				for _, v := range rule.Days {
					if day == v {
						applog.Printf("NextDate: найдено nextDate=%s для типа w", nextDate.Format(constants.TimeFormat))
						return nextDate.Format(constants.TimeFormat), nil
					}
				}
			}
		}
	case "m":
		if len(rule.Days) == 0 {
			applog.Printf("NextDate: пустой список дней для типа m")
			return "", fmt.Errorf("пустой список дней")
		}

		allowed := make(map[int]bool)
		if len(rule.Months) == 0 {
			for m := 1; m <= 12; m++ {
				allowed[m] = true
			}
		} else {
			for _, m := range rule.Months {
				if m >= 1 && m <= 12 {
					allowed[m] = true
				}
			}
		}
		if len(allowed) == 0 {
			applog.Printf("NextDate: нет разрешённых месяцев")
			return "", fmt.Errorf("нет разрешённых месяцев")
		}

		for {
			nextDate = nextDate.AddDate(0, 0, 1)
			if afterNow(nextDate, now) {
				break
			}
		}

		year := nextDate.Year()
		month := int(nextDate.Month())
		maxYear := year + 1

		var best time.Time
		bestSet := false

		for year <= maxYear {
			if allowed[month] {
				candidates := generateCandidatesM(year, time.Month(month), rule.Days, now.Location())
				for _, cand := range candidates {
					if cand.Before(nextDate) {
						continue
					}

					if !bestSet || cand.Before(best) {
						best = cand
						bestSet = true
					}
				}
			}

			month++
			if month > 12 {
				month = 1
				year++
			}

			if bestSet {
				applog.Printf("NextDate: найдено nextDate=%s для типа m", best.Format(constants.TimeFormat))
				return best.Format(constants.TimeFormat), nil
			}
		}
		if !bestSet {
			applog.Printf("NextDate: за год подходящая дата не найдена")
			return "", fmt.Errorf("за год подходящая дата не найдена")
		}
		return best.Format(constants.TimeFormat), nil
	case "y":
		for {
			nextDate = nextDate.AddDate(1, 0, 0)
			if afterNow(nextDate, now) {
				applog.Printf("NextDate: найдено nextDate=%s для типа y", nextDate.Format(constants.TimeFormat))
				return nextDate.Format(constants.TimeFormat), nil
			}
		}
	default:
		applog.Printf("NextDate: неизвестный тип правила %s", rule.Type)
		return "", errors.New("неизвестный тип правила")

	}

}

func afterNow(date, now time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()
	return time.Date(y1, m1, d1, 0, 0, 0, 0, now.Location()).After(time.Date(y2, m2, d2, 0, 0, 0, 0, now.Location()))
}

func generateCandidatesM(year int, month time.Month, days []int, loc *time.Location) []time.Time {
	var res []time.Time

	for _, d := range days {
		var cand time.Time
		firstDayNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
		switch d {
		case -1:
			cand = firstDayNextMonth.AddDate(0, 0, -1)
		case -2:
			cand = firstDayNextMonth.AddDate(0, 0, -2)
		default:
			if d < 1 || d > 31 {
				continue
			}
			cand = time.Date(year, month, d, 0, 0, 0, 0, loc)
			if cand.Month() != month {
				continue
			}
		}

		res = append(res, cand)
	}

	return res
}

func ruWeekDay(day time.Weekday) int {
	if day == time.Sunday {
		return 7
	}
	return int(day)

}
