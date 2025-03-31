package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
)

// AES-256 키 길이 상수
const KeySize = 32

// Encrypt 문자열과 키를 받아 암호화된 문자열을 반환
func Encrypt(content string, key []byte) (string, error) {
	if len(key) == 0 || len(content) == 0 {
		return "", errors.New("key와 content는 비어있을 수 없습니다")
	}

	// AES-256 키는 32바이트여야 함
	if len(key) != KeySize {
		return "", fmt.Errorf("키 길이가 %d바이트여야 함: 현재 %d 바이트", KeySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("AES 블록 생성 실패: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM 모드 초기화 실패: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce 생성 실패: %w", err)
	}

	// 암호화 실행
	ciphertext := gcm.Seal(nonce, nonce, []byte(content), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 암호화된 문자열과 키를 받아 원본 문자열을 반환
func Decrypt(encryptedText string, key []byte) (string, error) {
	if len(key) == 0 || len(encryptedText) == 0 {
		return "", errors.New("encryptedText와 key는 비어있을 수 없습니다")
	}

	// AES-256 키는 32바이트여야 함
	if len(key) != KeySize {
		return "", fmt.Errorf("키 길이가 %d바이트여야 함: 현재 %d 바이트", KeySize, len(key))
	}

	// Base64 디코딩
	cipherText, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("Base64 디코딩 실패: %w", err)
	}

	// AES 블록 생성
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("AES 블록 생성 실패: %w", err)
	}

	// GCM 모드 사용
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM 모드 초기화 실패: %w", err)
	}

	// Nonce 크기 확인
	if len(cipherText) < gcm.NonceSize() {
		return "", errors.New("암호화된 텍스트가 너무 짧습니다 (nonce 크기보다 작음)")
	}

	// Nonce 분리
	nonce, cipherTextWithoutNonce := cipherText[:gcm.NonceSize()], cipherText[gcm.NonceSize():]

	// 복호화 실행
	plainText, err := gcm.Open(nil, nonce, cipherTextWithoutNonce, nil)
	if err != nil {
		return "", fmt.Errorf("복호화 실패 (키가 올바르지 않거나 데이터가 변조됨): %w", err)
	}

	return string(plainText), nil
}

// CreateKeyFromString 임의의 문자열을 32바이트로 변환해 키로 사용
func CreateKeyFromString(input string) []byte {
	if len(input) == 0 {
		return nil // 빈 입력에 대한 처리
	}
	hash := sha256.Sum256([]byte(input))
	return hash[:]
}

// EncryptionResult 암호화 결과를 저장하는 구조체
type EncryptionResult struct {
	Original  string // 원본 문자열
	Encrypted string // 암호화된 문자열
	Error     error  // 발생한 오류 (있는 경우)
}

// DecryptionResult 복호화 결과를 저장하는 구조체
type DecryptionResult struct {
	Encrypted string // 암호화된 문자열
	Decrypted string // 복호화된 문자열
	Error     error  // 발생한 오류 (있는 경우)
}

// BulkEncrypt 여러 문자열을 동시에 암호화하는 함수
func BulkEncrypt(contents []string, key []byte) []EncryptionResult {
	if len(key) != KeySize {
		// 키 길이가 잘못된 경우 모든 결과에 동일한 오류 반환
		results := make([]EncryptionResult, len(contents))
		err := fmt.Errorf("키 길이가 %d바이트여야 함: 현재 %d 바이트", KeySize, len(key))
		for i, content := range contents {
			results[i] = EncryptionResult{
				Original: content,
				Error:    err,
			}
		}
		return results
	}

	results := make([]EncryptionResult, len(contents))

	// 동시에 처리할 고루틴 관리를 위한 WaitGroup
	var wg sync.WaitGroup

	// 각 문자열에 대해 고루틴 실행
	for i, content := range contents {
		wg.Add(1)

		go func(index int, text string) {
			defer wg.Done()

			// 빈 문자열 체크
			if len(text) == 0 {
				results[index] = EncryptionResult{
					Original: text,
					Error:    errors.New("content는 비어있을 수 없습니다"),
				}
				return
			}

			// 암호화 실행
			encrypted, err := Encrypt(text, key)

			// 결과 저장
			results[index] = EncryptionResult{
				Original:  text,
				Encrypted: encrypted,
				Error:     err,
			}
		}(i, content)
	}

	// 모든 고루틴이 완료될 때까지 대기
	wg.Wait()

	return results
}

// BulkDecrypt 여러 암호화된 문자열을 동시에 복호화하는 함수
func BulkDecrypt(encryptedTexts []string, key []byte) []DecryptionResult {
	if len(key) != KeySize {
		// 키 길이가 잘못된 경우 모든 결과에 동일한 오류 반환
		results := make([]DecryptionResult, len(encryptedTexts))
		err := fmt.Errorf("키 길이가 %d바이트여야 함: 현재 %d 바이트", KeySize, len(key))
		for i, text := range encryptedTexts {
			results[i] = DecryptionResult{
				Encrypted: text,
				Error:     err,
			}
		}
		return results
	}

	results := make([]DecryptionResult, len(encryptedTexts))

	// 동시에 처리할 고루틴 관리를 위한 WaitGroup
	var wg sync.WaitGroup

	// 각 암호화된 문자열에 대해 고루틴 실행
	for i, encryptedText := range encryptedTexts {
		wg.Add(1)

		go func(index int, text string) {
			defer wg.Done()

			// 빈 문자열 체크
			if len(text) == 0 {
				results[index] = DecryptionResult{
					Encrypted: text,
					Error:     errors.New("encryptedText는 비어있을 수 없습니다"),
				}
				return
			}

			// 복호화 실행
			decrypted, err := Decrypt(text, key)

			// 결과 저장
			results[index] = DecryptionResult{
				Encrypted: text,
				Decrypted: decrypted,
				Error:     err,
			}
		}(i, encryptedText)
	}

	// 모든 고루틴이 완료될 때까지 대기
	wg.Wait()

	return results
}

// BulkEncryptWithConcurrencyLimit 동시성 제한이 있는 대량 암호화 함수
func BulkEncryptWithConcurrencyLimit(contents []string, key []byte, concurrencyLimit int) []EncryptionResult {
	if concurrencyLimit <= 0 {
		concurrencyLimit = 10 // 기본 동시성 제한
	}

	if len(key) != KeySize {
		// 키 길이가 잘못된 경우 모든 결과에 동일한 오류 반환
		results := make([]EncryptionResult, len(contents))
		err := fmt.Errorf("키 길이가 %d바이트여야 함: 현재 %d 바이트", KeySize, len(key))
		for i, content := range contents {
			results[i] = EncryptionResult{
				Original: content,
				Error:    err,
			}
		}
		return results
	}

	results := make([]EncryptionResult, len(contents))

	// 작업 채널 생성
	jobs := make(chan int, len(contents))

	// 작업 채널에 인덱스 추가
	for i := range contents {
		jobs <- i
	}
	close(jobs)

	// 동시에 처리할 고루틴 관리를 위한 WaitGroup
	var wg sync.WaitGroup

	// 제한된 수의 워커 생성
	for w := 0; w < concurrencyLimit; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 작업 채널에서 인덱스를 가져와 처리
			for i := range jobs {
				content := contents[i]

				// 빈 문자열 체크
				if len(content) == 0 {
					results[i] = EncryptionResult{
						Original: content,
						Error:    errors.New("content는 비어있을 수 없습니다"),
					}
					continue
				}

				// 암호화 실행
				encrypted, err := Encrypt(content, key)

				// 결과 저장
				results[i] = EncryptionResult{
					Original:  content,
					Encrypted: encrypted,
					Error:     err,
				}
			}
		}()
	}

	// 모든 고루틴이 완료될 때까지 대기
	wg.Wait()

	return results
}

// BulkDecryptWithConcurrencyLimit 동시성 제한이 있는 대량 복호화 함수
func BulkDecryptWithConcurrencyLimit(encryptedTexts []string, key []byte, concurrencyLimit int) []DecryptionResult {
	if concurrencyLimit <= 0 {
		concurrencyLimit = 10 // 기본 동시성 제한
	}

	if len(key) != KeySize {
		// 키 길이가 잘못된 경우 모든 결과에 동일한 오류 반환
		results := make([]DecryptionResult, len(encryptedTexts))
		err := fmt.Errorf("키 길이가 %d바이트여야 함: 현재 %d 바이트", KeySize, len(key))
		for i, text := range encryptedTexts {
			results[i] = DecryptionResult{
				Encrypted: text,
				Error:     err,
			}
		}
		return results
	}

	results := make([]DecryptionResult, len(encryptedTexts))

	// 작업 채널 생성
	jobs := make(chan int, len(encryptedTexts))

	// 작업 채널에 인덱스 추가
	for i := range encryptedTexts {
		jobs <- i
	}
	close(jobs)

	// 동시에 처리할 고루틴 관리를 위한 WaitGroup
	var wg sync.WaitGroup

	// 제한된 수의 워커 생성
	for w := 0; w < concurrencyLimit; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 작업 채널에서 인덱스를 가져와 처리
			for i := range jobs {
				encryptedText := encryptedTexts[i]

				// 빈 문자열 체크
				if len(encryptedText) == 0 {
					results[i] = DecryptionResult{
						Encrypted: encryptedText,
						Error:     errors.New("encryptedText는 비어있을 수 없습니다"),
					}
					continue
				}

				// 복호화 실행
				decrypted, err := Decrypt(encryptedText, key)

				// 결과 저장
				results[i] = DecryptionResult{
					Encrypted: encryptedText,
					Decrypted: decrypted,
					Error:     err,
				}
			}
		}()
	}

	// 모든 고루틴이 완료될 때까지 대기
	wg.Wait()

	return results
}
