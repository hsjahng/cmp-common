package sql

import (
	"fmt"
	"github.com/hsjahng/cmp-common/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	gormLogger "gorm.io/gorm/logger"
	"sync"
	"testing"
	"time"
)

func Test_GetDB(t *testing.T) {
	logger.InitLogger(zapcore.InfoLevel.String())
	db, err := GetDB(DB_COMMON, gormLogger.Info)
	require.NoError(t, err)
	mariaDB, err := db.DB()
	require.NoError(t, err)
	mariaDB.SetMaxIdleConns(12)
	mariaDB.SetMaxOpenConns(12)
	mariaDB.SetConnMaxLifetime(time.Hour)
	mariaDB.SetConnMaxIdleTime(30 * time.Minute)

	// 2. Pool 설정이 제대로 적용되었는지 테스트
	t.Run("Connection Pool Settings", func(t *testing.T) {
		stats := mariaDB.Stats()
		assert.Equal(t, stats.MaxOpenConnections, 12)
		assert.GreaterOrEqual(t, stats.Idle, 1)
	})

	// 3. 실제 병렬 연결 테스트
	t.Run("Parallel Connections", func(t *testing.T) {
		var wg sync.WaitGroup
		connections := 20 // 병렬로 생성할 연결 수

		errChan := make(chan error, connections)

		// 동시에 여러 연결 시도
		for i := 0; i < connections; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// 간단한 쿼리 실행으로 연결 테스트
				var result int
				err := db.Raw("SELECT 1").Scan(&result).Error
				if err != nil {
					errChan <- fmt.Errorf("connection %d failed: %v", id, err)
					return
				}

				// 부하 테스트를 위한 짧은 대기 (선택적)
				time.Sleep(100 * time.Millisecond)

				// 또 다른 쿼리 실행 (연결이 재사용되는지 확인)
				err = db.Raw("SELECT 2").Scan(&result).Error
				if err != nil {
					errChan <- fmt.Errorf("connection %d second query failed: %v", id, err)
				}
			}(i)
		}

		// 모든 고루틴 완료 대기
		wg.Wait()
		close(errChan)

		// 에러 확인
		for err := range errChan {
			t.Error(err)
		}

		// 결과 통계 출력
		stats := mariaDB.Stats()
		t.Logf("After parallel test - Connection Stats: %+v", stats)

		// 검증: 연결 수가 우리가 설정한 제한 내에 있는지
		assert.LessOrEqual(t, stats.OpenConnections, 100, "Open connections should not exceed max")
		assert.True(t, stats.InUse <= stats.OpenConnections, "In-use connections should be <= open connections")
	})

	// 4. 연결 유지 테스트 (선택적)
	t.Run("Connection Keep-Alive", func(t *testing.T) {
		// 초기 상태 측정
		initialStats := mariaDB.Stats()
		t.Logf("Initial stats: %+v", initialStats)

		// 간단한 쿼리 실행
		var result int
		err := db.Raw("SELECT 1").Scan(&result).Error
		require.NoError(t, err)

		// 짧은 대기 후 통계 다시 확인 (연결이 계속 유지되는지)
		time.Sleep(1 * time.Second)
		afterStats := mariaDB.Stats()
		t.Logf("After query stats: %+v", afterStats)

		// 연결이 재사용되는지 확인
		assert.GreaterOrEqual(t, afterStats.Idle, 1, "Should have at least one idle connection")
	})

	// 테스트 종료시 리소스 정리
	t.Cleanup(func() {
		mariaDB.Close()
	})
}
