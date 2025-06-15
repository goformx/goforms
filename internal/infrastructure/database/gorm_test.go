package database

import (
	"context"
	"testing"
	"time"

	mocklogging "github.com/goformx/goforms/test/mocks/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// testTickerDuration is used for testing
const testTickerDuration = 10 * time.Millisecond

// testGormDB wraps GormDB for testing
type testGormDB struct {
	*GormDB
	tickerDuration time.Duration
}

// MonitorConnectionPool overrides the original method for testing
func (db *testGormDB) MonitorConnectionPool(ctx context.Context) {
	db.logger.Debug("starting MonitorConnectionPool")
	ticker := time.NewTicker(db.tickerDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			db.logger.Debug("MonitorConnectionPool context done")
			return
		case <-ticker.C:
			db.logger.Debug("MonitorConnectionPool tick")
			db.collectAndLogMetrics()
		}
	}
}

func TestMonitorConnectionPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocklogging.NewMockLogger(ctrl)

	// Set up expectations for the logger
	mockLogger.EXPECT().
		Debug(gomock.Any(), gomock.Any()).
		AnyTimes()

	mockLogger.EXPECT().
		Info("database connection pool status", gomock.Any()).
		Times(1)

	// Create a GORM DB instance with a mock configuration
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DriverName: "postgres",
		DSN:        "mock",
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Create a test GormDB instance with the mock DB
	db := &testGormDB{
		GormDB: &GormDB{
			DB:     gormDB,
			logger: mockLogger,
		},
		tickerDuration: testTickerDuration,
	}

	// Create a context that we can cancel
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start monitoring in a goroutine
	go db.MonitorConnectionPool(ctx)

	// Wait for the expected logger calls
	time.Sleep(50 * time.Millisecond)
}

func TestMonitorConnectionPoolWithHighUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocklogging.NewMockLogger(ctrl)

	// Set up expectations for the logger
	mockLogger.EXPECT().
		Debug(gomock.Any(), gomock.Any()).
		AnyTimes()

	mockLogger.EXPECT().
		Info("database connection pool status", gomock.Any()).
		Times(1)

	mockLogger.EXPECT().
		Warn("database connection pool usage is high", gomock.Any()).
		Times(1)

	// Create a GORM DB instance with a mock configuration
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DriverName: "postgres",
		DSN:        "mock",
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Create a test GormDB instance with the mock DB
	db := &testGormDB{
		GormDB: &GormDB{
			DB:     gormDB,
			logger: mockLogger,
		},
		tickerDuration: testTickerDuration,
	}

	// Create a context that we can cancel
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start monitoring in a goroutine
	go db.MonitorConnectionPool(ctx)

	// Wait for the expected logger calls
	time.Sleep(50 * time.Millisecond)
}

func TestMonitorConnectionPoolWithLongWait(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocklogging.NewMockLogger(ctrl)

	// Set up expectations for the logger
	mockLogger.EXPECT().
		Debug(gomock.Any(), gomock.Any()).
		AnyTimes()

	mockLogger.EXPECT().
		Info("database connection pool status", gomock.Any()).
		Times(1)

	mockLogger.EXPECT().
		Warn("database connection wait time is high", gomock.Any()).
		Times(1)

	// Create a GORM DB instance with a mock configuration
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DriverName: "postgres",
		DSN:        "mock",
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Create a test GormDB instance with the mock DB
	db := &testGormDB{
		GormDB: &GormDB{
			DB:     gormDB,
			logger: mockLogger,
		},
		tickerDuration: testTickerDuration,
	}

	// Create a context that we can cancel
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start monitoring in a goroutine
	go db.MonitorConnectionPool(ctx)

	// Wait for the expected logger calls
	time.Sleep(50 * time.Millisecond)
}
