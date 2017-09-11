package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	t.Log("it does not retryies if no error")
	{
		retryCnt := 0

		err := Times(2).Try(func(attempt uint) error {
			retryCnt++
			return nil
		})

		require.NoError(t, err)
		require.Equal(t, 1, retryCnt)
	}

	t.Log("it does retry if error")
	{
		attemptCnt := 0
		err := Times(2).Try(func(attempt uint) error {
			attemptCnt++
			return errors.New("error")
		})

		require.Error(t, err)
		require.Equal(t, "error", err.Error())
		require.Equal(t, 3, attemptCnt)
	}

	t.Log("it does not retry if Times=0")
	{
		attemptCnt := 0

		err := Times(0).Try(func(attempt uint) error {
			attemptCnt++
			return errors.New("error")
		})

		require.Error(t, err)
		require.Equal(t, "error", err.Error())
		require.Equal(t, 1, attemptCnt)
	}

	t.Log("it does a total attempt of 2 if Times=1")
	{
		attemptCnt := 0

		err := Times(1).Try(func(attempt uint) error {
			attemptCnt++
			return errors.New("error")
		})

		require.Error(t, err)
		require.Equal(t, "error", err.Error())
		require.Equal(t, 2, attemptCnt)
	}

	t.Log("it does a total attempt of 5 if Times=4")
	{
		attemptCnt := 0

		err := Times(4).Try(func(attempt uint) error {
			attemptCnt++
			return errors.New("error")
		})

		require.Error(t, err)
		require.Equal(t, "error", err.Error())
		require.Equal(t, 5, attemptCnt)
	}

	t.Log("it does not wait before first execution")
	{
		attemptCnt := 0
		startTime := time.Now()

		err := Times(1).Wait(3 * time.Second).Try(func(attempt uint) error {
			attemptCnt++
			return errors.New("error")
		})

		duration := time.Now().Sub(startTime)

		require.Error(t, err)
		require.Equal(t, "error", err.Error())
		require.Equal(t, 2, attemptCnt)
		if duration >= time.Duration(4)*time.Second {
			t.Fatalf("Should take no more than 4 sec, but got: %s", duration)
		}
	}

	t.Log("it waits before second execution")
	{
		attemptCnt := 0
		startTime := time.Now()

		err := Times(1).Wait(4 * time.Second).Try(func(attempt uint) error {
			attemptCnt++
			return errors.New("error")
		})

		duration := time.Now().Sub(startTime)

		require.Error(t, err)
		require.Equal(t, "error", err.Error())
		require.Equal(t, 2, attemptCnt)
		if duration < time.Duration(3)*time.Second {
			t.Fatalf("Should take at least 3 sec, but got: %s", duration)
		}
	}
}

func TestWait(t *testing.T) {
	t.Log("it creates retry model with wait time")
	{
		helper := Wait(3 * time.Second)
		require.Equal(t, 3*time.Second, helper.waitTime)
	}

	t.Log("it creates retry model with wait time")
	{
		helper := Wait(3 * time.Second)
		helper.Wait(5 * time.Second)
		require.Equal(t, 5*time.Second, helper.waitTime)
	}
}

func TestTimes(t *testing.T) {
	t.Log("it creates retry model with retry times")
	{
		helper := Times(3)
		require.Equal(t, uint(3), helper.retry)
	}

	t.Log("it sets retry times")
	{
		helper := Times(3)
		helper.Times(5)
		require.Equal(t, uint(5), helper.retry)
	}
}
