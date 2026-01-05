package luhn

// Validate проверяет номер по алгоритму Луна.
// Возвращает true, если номер проходит проверку.
func Validate(number string) bool {
	if len(number) == 0 {
		return false
	}

	var sum int
	parity := len(number) % 2

	for i, r := range number {
		if r < '0' || r > '9' {
			return false
		}

		digit := int(r - '0')

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}

// IsValidOrderNumber проверяет, является ли строка валидным номером заказа.
// Номер должен состоять только из цифр и проходить проверку по алгоритму Луна.
func IsValidOrderNumber(number string) bool {
	if len(number) == 0 {
		return false
	}

	for _, r := range number {
		if r < '0' || r > '9' {
			return false
		}
	}

	return Validate(number)
}
