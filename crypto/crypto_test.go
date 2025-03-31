package crypto

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func Test_Get(t *testing.T) {
	message := "암호화 해야하는 메시지"
	key := "Okestro2018@"
	byteKey := CreateKeyFromString(key)

	encrypted, err := Encrypt(message, byteKey)
	require.NoError(t, err)

	t.Log("암호화: ", encrypted)

	decrypted, err := Decrypt(encrypted, byteKey)
	require.NoError(t, err)

	t.Log("복호화: ", decrypted)
	require.Equal(t, message, decrypted)
}

func TestBulkEncryptDecrypt(t *testing.T) {
	// 테스트용 키 생성
	key := CreateKeyFromString("test-secret-key-for-bulk-operations")

	// 테스트할 원본 데이터
	originals := []string{
		"안녕하세요",
		"Hello, World!",
		"특수문자 !@#$%^&*()",
		"긴 문자열 테스트입니다. 이 문자열은 조금 더 깁니다. 대량 처리 시 문제가 없는지 확인하기 위한 테스트 데이터입니다.",
		"", // 빈 문자열 테스트
	}

	// 1. BulkEncrypt 테스트
	t.Run("BulkEncrypt", func(t *testing.T) {
		results := BulkEncrypt(originals, key)

		// 결과 갯수 확인
		if len(results) != len(originals) {
			t.Errorf("결과 갯수가 일치하지 않음: 원본 %d개, 결과 %d개", len(originals), len(results))
		}

		// 각 결과 검증
		for i, result := range results {
			// 원본 데이터 저장 확인
			if result.Original != originals[i] {
				t.Errorf("인덱스 %d: 원본 데이터가 일치하지 않음", i)
			}

			// 빈 문자열인 경우 오류 확인
			if originals[i] == "" {
				if result.Error == nil {
					t.Errorf("인덱스 %d: 빈 문자열에 대한 오류가 발생하지 않음", i)
				}
				continue
			}

			// 오류 없음 확인
			if result.Error != nil {
				t.Errorf("인덱스 %d: 암호화 오류 발생: %v", i, result.Error)
				continue
			}

			// 암호화된 문자열이 원본과 다른지 확인
			if result.Encrypted == originals[i] {
				t.Errorf("인덱스 %d: 암호화 실패 - 원본과 결과가 동일함", i)
			}

			// 암호화된 문자열이 비어있지 않은지 확인
			if result.Encrypted == "" {
				t.Errorf("인덱스 %d: 암호화 결과가 빈 문자열임", i)
			}
		}
	})

	// 2. BulkDecrypt 테스트 (위에서 암호화한 결과 사용)
	t.Run("BulkDecrypt", func(t *testing.T) {
		// 암호화 결과에서 암호문만 추출
		encryptedTexts := make([]string, len(originals))
		for i, result := range BulkEncrypt(originals, key) {
			if result.Error == nil {
				encryptedTexts[i] = result.Encrypted
			} else {
				encryptedTexts[i] = "" // 암호화 실패한 경우 빈 문자열
			}
		}

		// 복호화 실행
		decryptResults := BulkDecrypt(encryptedTexts, key)

		// 결과 갯수 확인
		if len(decryptResults) != len(encryptedTexts) {
			t.Errorf("복호화 결과 갯수가 일치하지 않음: 암호문 %d개, 결과 %d개",
				len(encryptedTexts), len(decryptResults))
		}

		// 각 결과 검증
		for i, result := range decryptResults {
			// 빈 문자열인 경우 오류 확인
			if encryptedTexts[i] == "" {
				if result.Error == nil {
					t.Errorf("인덱스 %d: 빈 암호문에 대한 오류가 발생하지 않음", i)
				}
				continue
			}

			// 오류 없음 확인
			if result.Error != nil {
				t.Errorf("인덱스 %d: 복호화 오류 발생: %v", i, result.Error)
				continue
			}

			// 복호화된 문자열이 원본과 일치하는지 확인
			if originals[i] != "" && result.Decrypted != originals[i] {
				t.Errorf("인덱스 %d: 복호화 결과가 원본과 일치하지 않음. 원본: %s, 복호화: %s",
					i, originals[i], result.Decrypted)
			}
		}
	})

	// 3. 동시성 제한 버전 테스트
	t.Run("BulkEncryptWithConcurrencyLimit", func(t *testing.T) {
		// 제한된 동시성으로 암호화
		results := BulkEncryptWithConcurrencyLimit(originals, key, 2)

		// 결과 갯수 확인
		if len(results) != len(originals) {
			t.Errorf("결과 갯수가 일치하지 않음: 원본 %d개, 결과 %d개", len(originals), len(results))
		}

		// 결과 검증 (빈 문자열 외에 모든 항목이 정상적으로 암호화되었는지)
		for i, result := range results {
			if originals[i] != "" && result.Error != nil {
				t.Errorf("인덱스 %d: 동시성 제한 버전 암호화 오류: %v", i, result.Error)
			}
		}
	})

	// 4. 잘못된 키 테스트
	t.Run("InvalidKey", func(t *testing.T) {
		invalidKey := []byte("too-short-key") // 키 길이가 너무 짧음

		results := BulkEncrypt(originals, invalidKey)

		// 모든 결과에 키 길이 오류가 있는지 확인
		for i, result := range results {
			if result.Error == nil {
				t.Errorf("인덱스 %d: 잘못된 키에 대한 오류가 발생하지 않음", i)
			}
		}
	})

	// 5. 라운드 트립 테스트 (암호화 후 복호화)
	t.Run("RoundTrip", func(t *testing.T) {
		// 빈 문자열을 제외한 테스트 데이터
		validData := []string{}
		for _, s := range originals {
			if s != "" {
				validData = append(validData, s)
			}
		}

		// 암호화
		encResults := BulkEncrypt(validData, key)

		// 암호문만 추출
		encryptedTexts := make([]string, len(encResults))
		for i, result := range encResults {
			encryptedTexts[i] = result.Encrypted
		}

		// 복호화
		decResults := BulkDecrypt(encryptedTexts, key)

		// 복호화된 결과만 추출
		decryptedTexts := make([]string, len(decResults))
		for i, result := range decResults {
			decryptedTexts[i] = result.Decrypted
		}

		// 원본과 복호화 결과 비교
		if !reflect.DeepEqual(validData, decryptedTexts) {
			t.Errorf("원본과 복호화 결과가 일치하지 않음\n원본: %v\n복호화: %v",
				validData, decryptedTexts)
		}
	})
}

// 벤치마크: 일반 버전 vs 동시성 제한 버전
func BenchmarkBulkEncrypt(b *testing.B) {
	// 테스트용 키와 대량의 데이터 준비
	key := CreateKeyFromString("benchmark-key")
	data := make([]string, 100)
	for i := 0; i < 100; i++ {
		data[i] = "벤치마크 테스트용 데이터 " + string(rune('A'+i%26))
	}

	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BulkEncrypt(data, key)
		}
	})

	b.Run("WithConcurrencyLimit-5", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BulkEncryptWithConcurrencyLimit(data, key, 5)
		}
	})

	b.Run("WithConcurrencyLimit-10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BulkEncryptWithConcurrencyLimit(data, key, 10)
		}
	})

	b.Run("WithConcurrencyLimit-20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BulkEncryptWithConcurrencyLimit(data, key, 20)
		}
	})
}
