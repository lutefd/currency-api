package logger

import (
	"context"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/repository"
	"github.com/robfig/cron/v3"
)

type PartitionManager struct {
	repo repository.LogRepository
	cron *cron.Cron
}

func NewPartitionManager(repo repository.LogRepository) *PartitionManager {
	c := cron.New()
	pm := &PartitionManager{
		repo: repo,
		cron: c,
	}

	_, err := c.AddFunc("0 0 1 * *", pm.createNextMonthPartitionWrapper)
	if err != nil {
		Errorf("failed to add cron job: %v\n", err)
	}

	return pm
}

func (pm *PartitionManager) Start(ctx context.Context) error {
	if err := pm.createInitialPartitions(ctx); err != nil {
		Errorf("failed to create initial partitions: %s", err)
	}

	pm.cron.Start()

	go func() {
		<-ctx.Done()
		pm.cron.Stop()
	}()

	return nil
}

func (pm *PartitionManager) createInitialPartitions(ctx context.Context) error {
	now := time.Now()
	for i := 0; i < 3; i++ {
		month := now.AddDate(0, i, 0)
		if err := pm.repo.CreatePartition(ctx, month); err != nil {
			return err
		}
	}
	return nil
}

func (pm *PartitionManager) createNextMonthPartition(ctx context.Context) error {
	nextMonth := time.Now().AddDate(0, 3, 0)
	return pm.repo.CreatePartition(ctx, nextMonth)
}

func (pm *PartitionManager) createNextMonthPartitionWrapper() {
	ctx := context.Background()
	if err := pm.createNextMonthPartition(ctx); err != nil {
		Errorf("failed to create next month partition: %v\n", err)
	}
}
