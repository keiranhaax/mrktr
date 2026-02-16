package api

import (
	_ "embed"
	"encoding/json"
	"math"
	"regexp"
	"sort"
	"strings"
)

const (
	maxExpandTokens     = 3
	minExpandScore      = 0.34
	minExpandSeparation = 0.08
	maxSuggestions      = 6
)

var (
	//go:embed data/products.json
	embeddedProductCatalog []byte

	tokenPattern = regexp.MustCompile(`[a-z0-9]+`)
	stopWords    = map[string]struct{}{
		"and": {}, "the": {}, "for": {}, "with": {}, "new": {}, "used": {}, "edition": {},
	}
)

// ProductEntry represents one product and its alias terms.
type ProductEntry struct {
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Synonyms []string `json:"synonyms"`
}

type productDocument struct {
	entry         ProductEntry
	nameLower     string
	synonymsLower []string
	vector        map[string]float64
}

// ProductIndex provides local query expansion and suggestions.
type ProductIndex struct {
	products []productDocument
	idf      map[string]float64
}

// NewProductIndex builds a product index from embedded catalog data.
func NewProductIndex() *ProductIndex {
	entries := loadProductCatalog()
	return newProductIndexFromEntries(entries)
}

func newProductIndexFromEntries(entries []ProductEntry) *ProductIndex {
	idx := &ProductIndex{
		products: make([]productDocument, 0, len(entries)),
		idf:      map[string]float64{},
	}
	if len(entries) == 0 {
		return idx
	}

	documentTerms := make([]map[string]float64, 0, len(entries))
	df := map[string]int{}

	for _, raw := range entries {
		entry := sanitizeEntry(raw)
		if entry.Name == "" {
			continue
		}

		nameTokens := tokenize(entry.Name)
		synonymTokens := tokenize(strings.Join(entry.Synonyms, " "))
		categoryTokens := tokenize(entry.Category)
		if len(nameTokens) == 0 && len(synonymTokens) == 0 {
			continue
		}

		terms := map[string]float64{}
		for _, token := range nameTokens {
			terms[token] += 2.0
		}
		for _, token := range synonymTokens {
			terms[token] += 1.5
		}
		for _, token := range categoryTokens {
			terms[token] += 0.5
		}

		seen := map[string]struct{}{}
		for token := range terms {
			if _, ok := seen[token]; ok {
				continue
			}
			seen[token] = struct{}{}
			df[token]++
		}

		doc := productDocument{
			entry:     entry,
			nameLower: strings.ToLower(entry.Name),
		}
		for _, synonym := range entry.Synonyms {
			doc.synonymsLower = append(doc.synonymsLower, strings.ToLower(synonym))
		}

		idx.products = append(idx.products, doc)
		documentTerms = append(documentTerms, terms)
	}

	n := float64(len(documentTerms))
	for token, docsWithToken := range df {
		idx.idf[token] = math.Log((1.0+n)/(1.0+float64(docsWithToken))) + 1.0
	}

	for i, terms := range documentTerms {
		vector := weightedVector(terms, idx.idf)
		idx.products[i].vector = normalizeVector(vector)
	}

	return idx
}

// Expand rewrites vague queries into a best-fit product name when confidence is high.
func (idx *ProductIndex) Expand(query string) string {
	trimmed := strings.TrimSpace(query)
	if idx == nil || len(idx.products) == 0 || trimmed == "" {
		return trimmed
	}

	tokens := tokenize(trimmed)
	if len(tokens) == 0 || len(tokens) > maxExpandTokens {
		return trimmed
	}

	queryVector := normalizeVector(weightedVector(termCounts(tokens), idx.idf))
	if len(queryVector) == 0 {
		return trimmed
	}

	queryLower := strings.ToLower(trimmed)
	topName := ""
	topScore := 0.0
	secondScore := 0.0

	for _, product := range idx.products {
		score := cosineSimilarity(queryVector, product.vector)
		if queryLower == product.nameLower {
			score += 0.50
		}
		for _, synonym := range product.synonymsLower {
			if queryLower == synonym {
				score += 0.50
				break
			}
			if len(queryLower) >= 3 && strings.HasPrefix(synonym, queryLower) {
				score += 0.10
				break
			}
		}

		if score > topScore {
			secondScore = topScore
			topScore = score
			topName = product.entry.Name
			continue
		}
		if score > secondScore {
			secondScore = score
		}
	}

	if topName == "" {
		return trimmed
	}
	if topScore < minExpandScore {
		return trimmed
	}
	if (topScore - secondScore) < minExpandSeparation {
		return trimmed
	}
	if strings.EqualFold(trimmed, topName) {
		return trimmed
	}

	return topName
}

// Suggest returns ranked product suggestions for the current input prefix.
func (idx *ProductIndex) Suggest(prefix string) []string {
	p := strings.TrimSpace(strings.ToLower(prefix))
	if idx == nil || len(idx.products) == 0 || len(p) < 2 {
		return nil
	}

	queryVector := normalizeVector(weightedVector(termCounts(tokenize(p)), idx.idf))

	type candidate struct {
		value string
		score float64
	}
	candidates := make([]candidate, 0, len(idx.products))

	for _, product := range idx.products {
		baseScore := 0.0
		if len(queryVector) > 0 {
			baseScore = cosineSimilarity(queryVector, product.vector)
		}

		namePrefix := strings.HasPrefix(product.nameLower, p)
		nameTokenPrefix := tokenHasPrefix(product.nameLower, p)
		if namePrefix || nameTokenPrefix {
			score := 2.4 + baseScore
			if namePrefix {
				score = 3.0 + baseScore
			}
			if product.nameLower == p {
				score += 1.0
			}
			candidates = append(candidates, candidate{
				value: product.entry.Name,
				score: score,
			})
		}

		for i, synonym := range product.synonymsLower {
			synPrefix := strings.HasPrefix(synonym, p)
			synTokenPrefix := tokenHasPrefix(synonym, p)
			if !synPrefix && !synTokenPrefix {
				continue
			}

			score := 3.4 + baseScore
			if synPrefix {
				score = 4.0 + baseScore
			}
			if synonym == p {
				score += 1.0
			}
			candidates = append(candidates, candidate{
				value: product.entry.Synonyms[i],
				score: score,
			})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].value < candidates[j].value
		}
		return candidates[i].score > candidates[j].score
	})

	out := make([]string, 0, maxSuggestions)
	seen := map[string]struct{}{}
	for _, c := range candidates {
		key := strings.ToLower(strings.TrimSpace(c.value))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, c.value)
		if len(out) == maxSuggestions {
			break
		}
	}
	return out
}

func tokenHasPrefix(text, prefix string) bool {
	if prefix == "" {
		return false
	}
	for _, token := range tokenPattern.FindAllString(strings.ToLower(text), -1) {
		if strings.HasPrefix(token, prefix) {
			return true
		}
	}
	return false
}

func loadProductCatalog() []ProductEntry {
	if len(embeddedProductCatalog) == 0 {
		return defaultProductCatalog()
	}

	var entries []ProductEntry
	if err := json.Unmarshal(embeddedProductCatalog, &entries); err != nil {
		return defaultProductCatalog()
	}
	if len(entries) == 0 {
		return defaultProductCatalog()
	}
	return entries
}

func defaultProductCatalog() []ProductEntry {
	return []ProductEntry{
		{Name: "Nintendo Switch OLED", Category: "Gaming", Synonyms: []string{"switch", "nintendo switch"}},
		{Name: "PlayStation 5 Console", Category: "Gaming", Synonyms: []string{"ps5", "playstation 5"}},
		{Name: "AirPods Pro 2", Category: "Audio", Synonyms: []string{"airpods pro", "airpods"}},
	}
}

func sanitizeEntry(entry ProductEntry) ProductEntry {
	entry.Name = strings.TrimSpace(entry.Name)
	entry.Category = strings.TrimSpace(entry.Category)

	synonyms := make([]string, 0, len(entry.Synonyms))
	seen := map[string]struct{}{}
	for _, raw := range entry.Synonyms {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		synonyms = append(synonyms, trimmed)
	}
	entry.Synonyms = synonyms
	return entry
}

func tokenize(text string) []string {
	raw := tokenPattern.FindAllString(strings.ToLower(text), -1)
	if len(raw) == 0 {
		return nil
	}

	out := make([]string, 0, len(raw))
	for _, token := range raw {
		if _, skip := stopWords[token]; skip {
			continue
		}
		out = append(out, token)
	}
	return out
}

func termCounts(tokens []string) map[string]float64 {
	if len(tokens) == 0 {
		return nil
	}
	counts := make(map[string]float64, len(tokens))
	for _, token := range tokens {
		counts[token]++
	}
	return counts
}

func weightedVector(counts map[string]float64, idf map[string]float64) map[string]float64 {
	if len(counts) == 0 {
		return nil
	}
	vector := map[string]float64{}
	for term, tf := range counts {
		termIDF, ok := idf[term]
		if !ok {
			continue
		}
		vector[term] = (1.0 + math.Log(tf)) * termIDF
	}
	return vector
}

func normalizeVector(vector map[string]float64) map[string]float64 {
	if len(vector) == 0 {
		return nil
	}

	norm := 0.0
	for _, weight := range vector {
		norm += weight * weight
	}
	if norm == 0 {
		return nil
	}
	scale := 1.0 / math.Sqrt(norm)

	out := make(map[string]float64, len(vector))
	for term, weight := range vector {
		out[term] = weight * scale
	}
	return out
}

func cosineSimilarity(a, b map[string]float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	score := 0.0
	for term, left := range a {
		if right, ok := b[term]; ok {
			score += left * right
		}
	}
	return score
}
