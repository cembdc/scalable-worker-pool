package workerpool

import (
	"context"
	"fmt"
	"sync"
)

// Dispatcher yapısını dışa açık hale getir
type Dispatcher struct {
	inCh          chan Request
	workerManager *WorkerManager
	scaler        *Scaler
	reqHandler    map[int]RequestHandler
}

// NewDispatcher fonksiyonunun dönüş tipini *Dispatcher olarak değiştir
func NewDispatcher(
	bufferSize int,
	wg *sync.WaitGroup,
	maxWorkers int,
	reqHandler map[int]RequestHandler,
) WorkerPoolManager {
	inCh := make(chan Request, bufferSize)
	stopCh := make(chan struct{}, maxWorkers)
	workerManager := NewWorkerManager(wg, inCh, stopCh, reqHandler)
	scaler := NewScaler(workerManager, inCh, DefaultMinWorkers, DefaultMaxWorkers, DefaultLoadThreshold)

	return &Dispatcher{
		inCh:          inCh,
		workerManager: workerManager,
		scaler:        scaler,
		reqHandler:    reqHandler,
	}
}

func (d *Dispatcher) AddWorker(w *Worker) {
	d.workerManager.AddWorker(w)
}

func (d *Dispatcher) RemoveWorker(minWorkers int) {
	if d.workerManager.WorkerCount() > minWorkers {
		d.workerManager.RemoveWorker()
	}
}

func (d *Dispatcher) ScaleWorkers(ctx context.Context, minWorkers, maxWorkers, loadThreshold int) {
	d.scaler.Start(ctx)
}

func (d *Dispatcher) MakeRequest(r Request) {
	select {
	case d.inCh <- r:
	default:
		fmt.Println("Request channel is full. Dropping request.")
	}
}

func (d *Dispatcher) Stop(ctx context.Context) {
	fmt.Println("\nGraceful shutdown initiated")

	// Önce yeni isteklerin alınmasını durdur
	close(d.inCh)

	// Bekleyen isteklerin sayısını kontrol et
	pendingRequests := len(d.inCh)
	fmt.Printf("Pending requests: %d\n", pendingRequests)

	// Tüm işçileri durdur
	d.workerManager.StopAllWorkers()

	// İşçilerin tamamlanmasını bekle
	done := make(chan struct{})
	go func() {
		d.workerManager.WaitForAllWorkers()
		close(done)
	}()

	// Timeout veya tamamlanma için bekle
	select {
	case <-done:
		fmt.Println("All workers stopped gracefully")
	case <-ctx.Done():
		fmt.Println("Timeout reached, some requests may not have been processed")
	}

	// Kalan istekleri raporla
	remainingRequests := len(d.inCh)
	fmt.Printf("Unprocessed requests: %d\n", remainingRequests)

	// İsteğe bağlı: Kalan istekleri bir log dosyasına veya başka bir sisteme kaydet
	if remainingRequests > 0 {
		d.logRemainingRequests()
	}

	fmt.Println("Shutdown complete")
}

func (d *Dispatcher) logRemainingRequests() {
	for req := range d.inCh {
		fmt.Printf("Unprocessed request: %v\n", req)
		// Burada istekleri bir dosyaya veya veritabanına kaydedebilirsiniz
	}
}
