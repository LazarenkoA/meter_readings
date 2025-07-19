package mos_ru

import (
	"crypto/sha1"
	"errors"
	"strconv"
	"strings"
)

var alphabet = []byte("0123456789/+abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// инкрементирует счётчик; возвращает true, если получилось перейти к новому числу
func next(digits []byte) bool {
	for i := len(digits) - 1; i >= 0; i-- {
		if digits[i] < byte(len(alphabet)-1) {
			digits[i]++
			return true
		}
		digits[i] = 0
	}
	return false
}

func digitsToString(d []byte) string {
	out := make([]byte, len(d))
	for i, v := range d {
		out[i] = alphabet[v]
	}
	return string(out)
}

// leadingZeroBits возвращает количество ведущих нулевых битов в хэше
func leadingZeroBits(h []byte) int {
	bits := 0
	for _, b := range h {
		if b == 0 {
			bits += 8
			continue
		}
		for i := 7; i >= 0; i-- {
			if (b>>i)&1 == 0 {
				bits++
			} else {
				return bits
			}
		}
	}
	return bits
}

// solve подбирает nonce, чтобы SHA‑1(prefix+nonce) имел ≥ bits нулевых битов
func solve(prefix string, bits int) (string, error) {
	counter := make([]byte, 1) // начинаем с одной «цифры»
	for {
		header := prefix + digitsToString(counter)
		sum := sha1.Sum([]byte(header))

		if leadingZeroBits(sum[:]) >= bits {
			return digitsToString(counter), nil
		}

		if next(counter) {
			continue
		}
		if len(counter) >= 25 { // лимит из pOfw.js
			return "", errors.New("counter overflow (>25 chars)")
		}
		counter = append(make([]byte, len(counter)+1)) // увеличиваем разрядность
	}
}

// Build принимает исходное значение из <input id="pow"> и возвращает окончательный proofOfWork.
func buildPow(seed string) (string, error) {
	parts := strings.Split(seed, ":")
	if len(parts) < 3 {
		return "", errors.New("bad seed")
	}

	bits, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(seed, ":") {
		seed += ":"
	}

	nonce, err := solve(seed, bits)
	if err != nil {
		return "", err
	}
	return seed + nonce, nil
}
