package cardrender

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type ChromePool struct {
	browser *rod.Browser
	mu      sync.Mutex
}

func NewChromePool() (*ChromePool, error) {
	var l *launcher.Launcher

	if bin := os.Getenv("ROD_BROWSER_BIN"); bin != "" {
		l = launcher.New().Bin(bin)
	} else {
		l = launcher.New()
	}

	u, err := l.Headless(true).
		Set("disable-gpu").
		Set("no-sandbox").
		Set("disable-dev-shm-usage").
		Launch()
	if err != nil {
		return nil, fmt.Errorf("launch chrome: %w", err)
	}

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connect chrome: %w", err)
	}

	return &ChromePool{browser: browser}, nil
}

func (p *ChromePool) GetBrowser() *rod.Browser {
	return p.browser
}

func (p *ChromePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.browser != nil {
		return p.browser.Close()
	}
	return nil
}
