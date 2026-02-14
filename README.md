# mrktr

A terminal-based reseller price research tool for comparing marketplace prices.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Platform](https://img.shields.io/badge/Platform-macOS%20|%20Linux%20|%20Windows-blue)

## Preview

```
┌─ Search ─────────────────────────┐┌─ Statistics ───────────────────┐
│ > iPhone 14 Pro                  ││ Results: 8 listings            │
└──────────────────────────────────┘│                                │
┌─ Results ────────────────────────┐│ Min:     $680.00               │
│ Platform    Price   Cond  Status ││ Max:     $920.00               │
│ eBay       $760.00  Used  Sold   ││ Average: $800.00               │
│ eBay       $840.00  New   Active ││ Median:  $795.00               │
│ Mercari    $720.00  Good  Active │└────────────────────────────────┘
│ Mercari    $680.00  Used  Sold   │┌─ Profit Calculator ───────────┐
│ Amazon     $920.00  New   Active ││ Your Cost: $500                │
│ eBay       $736.00  Used  Sold   ││ ─────────────────────────────  │
│ Facebook   $640.00  Fair  Active ││ At Avg:  +$300.00 (60%)        │
│ eBay       $800.00  Good  Active ││ At Min:  +$180.00 (36%)        │
└──────────────────────────────────┘│ At Max:  +$420.00 (84%)        │
┌─ History ────────────────────────┴┴────────────────────────────────┐
│ Recent: iPhone 14 Pro | PS5 | Nintendo Switch | AirPods Pro        │
└────────────────────────────────────────────────────────────────────┘
 / search  j/k navigate  Enter select  Tab panels  c cost  q quit
```

## Features

- **Multi-Marketplace Search** - Compare prices across eBay, Mercari, Amazon, and Facebook Marketplace
- **Real-Time Statistics** - Instantly see min, max, average, and median prices
- **Profit Calculator** - Enter your cost and see potential profit margins
- **Search History** - Quick access to recent searches
- **Vim-Style Navigation** - Navigate with j/k keys or arrow keys
- **Clean Dashboard UI** - Professional panel-based interface

## Installation

### Prerequisites

- Go 1.21 or higher

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/mrktr.git
cd mrktr

# Build the binary
go build -o mrktr .

# Run
./mrktr
```

### Install to PATH

```bash
# After building, copy to your PATH
sudo cp mrktr /usr/local/bin/

# Now run from anywhere
mrktr
```

## Configuration

mrktr uses search APIs to fetch real marketplace data. Configure your API keys as environment variables:

```bash
# Firecrawl (primary)
export FIRECRAWL_API_KEY="your-firecrawl-api-key"

# Tavily (fallback)
export TAVILY_API_KEY="your-tavily-api-key"
```

### Demo Mode

If no API keys are configured, mrktr runs in demo mode with sample data. This is useful for testing the interface.

### Adding to Shell Profile

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
export FIRECRAWL_API_KEY="your-key-here"
export TAVILY_API_KEY="your-key-here"
```

## Usage

### Basic Workflow

1. **Launch the app**
   ```bash
   mrktr
   ```

2. **Search for an item**
   - Type your search query (e.g., "iPhone 14 Pro")
   - Press `Enter` to search

3. **Review results**
   - Use `j/k` or arrow keys to navigate results
   - View statistics in the right panel

4. **Calculate profit**
   - Press `c` to focus the calculator
   - Enter your cost
   - See profit margins at different price points

5. **Open listing**
   - Press `Enter` on a result to open the URL in your browser

## Keybindings

| Key | Action |
|-----|--------|
| `/` | Focus search input |
| `Enter` | Execute search / Open selected URL |
| `Tab` | Cycle between panels |
| `Shift+Tab` | Cycle panels backwards |
| `j` / `Down` | Move down in list |
| `k` / `Up` | Move up in list |
| `c` | Focus profit calculator |
| `Esc` | Unfocus current panel |
| `q` | Quit application |
| `Ctrl+C` | Force quit |

## Project Structure

```
mrktr/
├── main.go          # Entry point, program initialization
├── model.go         # Application state and data structures
├── update.go        # Keyboard handling and state updates
├── view.go          # UI rendering logic
├── styles.go        # Lip Gloss styles and colors
├── api.go           # Search API integration
├── types/           # Listing and statistics types
│   └── listing.go
├── go.mod           # Go module definition
└── README.md        # This file
```

## Tech Stack

- **Language:** [Go](https://golang.org/)
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Components:** [Bubbles](https://github.com/charmbracelet/bubbles)
- **Search APIs:** Firecrawl, Tavily

## How It Works

1. **Search Query** - User enters an item name
2. **API Request** - Query is sent to Firecrawl/Tavily with marketplace site filters
3. **Price Parsing** - Regex extracts prices from search results
4. **Platform Detection** - URLs are parsed to identify the marketplace
5. **Statistics** - Min, max, average, and median are calculated
6. **Display** - Results are rendered in the dashboard

## Contributing

Contributions are welcome! Here's how to get started:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development

```bash
# Run with live reload (requires air)
air

# Or run directly
go run .

# If your environment blocks the default Go cache location
GOCACHE=$(pwd)/.cache/go-build GOMODCACHE=$(pwd)/.cache/go-mod go test ./...
```

## Roadmap

- [ ] Persistent search history
- [ ] Export results to CSV
- [ ] Price alerts
- [ ] More marketplace support
- [ ] Saved searches
- [ ] Price history graphs

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
