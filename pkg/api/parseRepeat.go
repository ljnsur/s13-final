package api

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	applog "github.com/ljnsur/todosay/pkg/log"
)

type RepeatRule struct {
	Type   string
	Months []int
	Days   []int
}

// parseRepeatRule разбирает строковое правило повтора и возвращает структурированное представление.
// Ожидаемый формат строки: тип + аргументы, например:
//
//	d N         — повтор каждые N дней
//	w D1,D2,... — повтор по дням недели (1..7)
//	m D[,D2][ M1,M2 ] — повтор по дням месяца и опционально по месяцам
//	y           — ежегодно (годовой)
//
// Функция выполняет валидацию аргументов и возвращает подробные ошибки при некорректном вводе.
func parseRepeatRule(repeat string) (RepeatRule, error) {
	applog.Printf("parseRepeatRule: разбор правила %q", repeat)
	// Разбиваем строку на токены: тип и последующие аргументы
	parts := strings.Split(repeat, " ")

	// Защитный код: если строка пустая или нет токенов — сообщаем об ошибке
	if len(parts) == 0 || parts[0] == "" {
		applog.Printf("parseRepeatRule: пустая строка правила")
		return RepeatRule{}, errors.New("пустое правило повторения")
	}

	switch parts[0] {
	case "d":
		// Тип 'd' — ожидается ровно один аргумент: число дней (N)
		if len(parts) != 2 {
			applog.Printf("parseRepeatRule: неверное количество аргументов для d: got=%d", len(parts))
			return RepeatRule{}, errors.New("неверное количeство аргументов ")
		}
		// парсим количество дней и проверяем пределы
		day, err := strconv.Atoi(parts[1])
		if err != nil {
			applog.Printf("parseRepeatRule: нецелое число для d: %v", err)
			return RepeatRule{}, fmt.Errorf("аргумент %s нецелое число: %s ", parts[1], err)
		}
		if day < 1 || day > 400 {
			// Ограничение в 400 выставлено чтобы избежать бесконечных или крайне редких повторов
			applog.Printf("parseRepeatRule: недопустимое значение дней d=%d", day)
			return RepeatRule{}, errors.New("количство дней должно быть не менее 1 и не более 400 ")
		}
		// Возвращаем правило с одним значением в поле Days — удобный унифицированный формат
		var daysINT []int
		daysINT = append(daysINT, day)
		return RepeatRule{Type: parts[0], Days: daysINT}, nil
	case "w":
		// Тип 'w' — ожидаем список номеров дней недели, разделённых запятыми
		if len(parts) != 2 {
			applog.Printf("parseRepeatRule: неверное количество аргументов для w")
			return RepeatRule{}, errors.New("неверное количeство аргументов ")
		}
		days := strings.Split(parts[1], ",")

		var daysINT []int
		for _, v := range days {
			// Парсим каждый номер дня и проверяем, что он в диапазоне 1..7
			day, err := strconv.Atoi(v)
			if err != nil {
				applog.Printf("parseRepeatRule: нецелое число в w: %v", err)
				return RepeatRule{}, fmt.Errorf("аргумент %d нецелое число: %s ", day, err)
			}
			if day < 1 || day > 7 {
				applog.Printf("parseRepeatRule: недопустимый день w=%d", day)
				return RepeatRule{}, errors.New("номера дней должны быть с 1 по 7 включительно  ")
			}
			daysINT = append(daysINT, day)
		}

		// Сортируем, чтобы иметь каноническое представление правила (полезно для сравнения и логов)
		sort.Ints(daysINT)
		return RepeatRule{Type: parts[0], Days: daysINT}, nil
	case "m":
		// Тип 'm' — поддерживает два варианта:
		//  1) "m D1,D2,..." — только дни месяца
		//  2) "m D1,D2,... M1,M2,..." — дни месяца + ограничение по месяцам
		if len(parts) == 2 {
			days := strings.Split(parts[1], ",")

			var daysINT []int
			for _, v := range days {
				// День месяца может быть -1 (последний), -2 (предпоследний) или 1..31
				day, err := strconv.Atoi(v)
				if err != nil {
					applog.Printf("parseRepeatRule: нецелое число в m: %v", err)
					return RepeatRule{}, fmt.Errorf("аргумент %d нецелое число: %s ", day, err)
				}

				if !((day >= 1 && day <= 31) || day == -1 || day == -2) {
					applog.Printf("parseRepeatRule: недопустимый день m=%d", day)
					return RepeatRule{}, errors.New("номера дней должны быть с 1 по 31 включительно, либо = -1, -2 ")
				}
				daysINT = append(daysINT, day)
			}

			sort.Ints(daysINT) // упорядочим дни для удобства
			return RepeatRule{Type: parts[0], Days: daysINT}, nil

		} else if len(parts) == 3 {
			// Вариант с указанием месяцев: парсим дни, затем парсим список месяцев
			days := strings.Split(parts[1], ",")

			var daysINT []int
			for _, v := range days {
				day, err := strconv.Atoi(v)
				if err != nil {
					applog.Printf("parseRepeatRule: нецелое число в m: %v", err)
					return RepeatRule{}, fmt.Errorf("аргумент %d нецелое число: %s ", day, err)
				}
				if !((day >= 1 && day <= 31) || day == -1 || day == -2) {
					applog.Printf("parseRepeatRule: недопустимый день m=%d", day)
					return RepeatRule{}, errors.New("номера дней должны быть с 1 по 31 включительно, либо = -1, -2 ")
				}
				daysINT = append(daysINT, day)
			}
			months := strings.Split(parts[2], ",")

			var monthsINT []int
			for _, v := range months {
				// Парсим каждый номер месяца и проверяем диапазон 1..12
				months, err := strconv.Atoi(v)
				if err != nil {
					applog.Printf("parseRepeatRule: нецелое число месяца: %v", err)
					return RepeatRule{}, fmt.Errorf("аргумент %d нецелое число ", months)
				}
				if months < 1 || months > 12 {
					applog.Printf("parseRepeatRule: недопустимый месяц m=%d", months)
					return RepeatRule{}, errors.New("номера месяце должны быть с 1 по 12 включительно ")
				}
				monthsINT = append(monthsINT, months)
			}

			sort.Ints(daysINT)
			sort.Ints(monthsINT)
			// Возвращаем правило с днями и месяцами — это позволяет использовать одно и то же
			// представление при расчётах следующей даты.
			return RepeatRule{Type: parts[0], Days: daysINT, Months: monthsINT}, nil

		} else {
			applog.Printf("parseRepeatRule: невалидный формат аргументов для m")
			return RepeatRule{}, errors.New("невалидный формат аргументов для месяца(m) ")
		}
	case "y":
		// Тип 'y' (годовой) не требует аргументов — просто возвращаем маркировку типа
		applog.Printf("parseRepeatRule: правило типа y")
		return RepeatRule{Type: parts[0]}, nil
	default:
		applog.Printf("parseRepeatRule: неизвестный тип правила %s", parts[0])
		return RepeatRule{}, fmt.Errorf("неизвестный тип правила %s ", parts[0])

	}

}
