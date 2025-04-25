package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"

	"github.com/faisaloncode/ecommerce-crawler/crawler/models"
	"github.com/faisaloncode/ecommerce-crawler/crawler/config"
)

type CategoryScraper struct {
	db         *gorm.DB
	httpClient *http.Client
	baseURL    string
}

func NewCategoryScraper(db *gorm.DB, cfg *config.Config) *CategoryScraper {
	return &CategoryScraper{
		db:         db,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    cfg.BaseURL,
	}
}

// ScrapeCategories fetches all categories from Trendyol
func (s *CategoryScraper) ScrapeCategories() error {
	log.Println("Starting category scraping process")
	
	// Fetch the main page
	resp, err := s.httpClient.Get(s.baseURL)
	if err != nil {
		return fmt.Errorf("failed to fetch main page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status: %s", resp.Status)
	}

	// Parse HTML using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find main category navigation
	// Note: These selectors would need to be adjusted based on Trendyol's actual HTML structure
	mainCategories := doc.Find("nav.main-nav ul.main-menu > li")

	// Process each main category
	// Capture the scraper instance for use in closure
	scraper := s

	mainCategories.Each(func(i int, sel *goquery.Selection) {
		categoryName := strings.TrimSpace(sel.Find("a span").First().Text())
		categoryURL, exists := sel.Find("a").First().Attr("href")
		
		if exists && categoryName != "" {
			// Create or update main category
			mainCategory := models.Category{
				Name:      categoryName,
				Slug:      extractSlug(categoryURL),
				ExternalID: extractCategoryID(categoryURL),
			}
			
			result := scraper.db.Where("name = ?", mainCategory.Name).FirstOrCreate(&mainCategory)
			if result.Error != nil {
				log.Printf("Error saving main category %s: %v", mainCategory.Name, result.Error)
				return
			}
			
			// Process subcategories
			sel.Find("div.sub-menu .sub-item-list li").Each(func(j int, subSel *goquery.Selection) {
				subCategoryName := strings.TrimSpace(subSel.Find("a").Text())
				subCategoryURL, subExists := subSel.Find("a").Attr("href")
				
				if subExists && subCategoryName != "" {
					// Create or update subcategory
					subCategory := models.Category{
						Name:      subCategoryName,
						ParentID:  &mainCategory.ID,
						Slug:      extractSlug(subCategoryURL),
						ExternalID: extractCategoryID(subCategoryURL),
					}
					
					result := s.db.Where("name = ? AND parent_id = ?", subCategory.Name, mainCategory.ID).FirstOrCreate(&subCategory)
					if result.Error != nil {
						log.Printf("Error saving subcategory %s: %v", subCategory.Name, result.Error)
					}
				}
			})
		}
	})

	log.Println("Category scraping completed successfully")
	return nil
}

// Alternative approach: Use Trendyol's API if available
func (s *CategoryScraper) ScrapeCategoriesAPI() error {
	log.Println("Starting category scraping via API")
	
	// This would be the API endpoint for categories if available
	apiURL := fmt.Sprintf("%s/api/v1/categories", s.baseURL)
	
	resp, err := s.httpClient.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch categories API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad API response status: %s", resp.Status)
	}

	// Parse JSON response
	var categories struct {
		Data []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			ParentID string `json:"parentId"`
			Slug     string `json:"slug"`
			Children []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				ParentID string `json:"parentId"`
				Slug     string `json:"slug"`
			} `json:"children"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return fmt.Errorf("failed to decode API response: %w", err)
	}

	// Process categories from API
	for _, mainCat := range categories.Data {
		// Create or update main category
		var parentID *uint
		mainCategory := models.Category{
			ExternalID: mainCat.ID,
			Name:       mainCat.Name,
			Slug:       mainCat.Slug,
			ParentID:   parentID,
		}
		
		result := s.db.Where("external_id = ?", mainCategory.ExternalID).FirstOrCreate(&mainCategory)
		if result.Error != nil {
			log.Printf("Error saving main category %s: %v", mainCategory.Name, result.Error)
			continue
		}
		
		// Process subcategories
		for _, subCat := range mainCat.Children {
			subCategory := models.Category{
				ExternalID: subCat.ID,
				Name:       subCat.Name,
				Slug:       subCat.Slug,
				ParentID:   &mainCategory.ID,
			}
			
			result := s.db.Where("external_id = ?", subCategory.ExternalID).FirstOrCreate(&subCategory)
			if result.Error != nil {
				log.Printf("Error saving subcategory %s: %v", subCategory.Name, result.Error)
			}
		}
	}

	log.Println("Category scraping via API completed successfully")
	return nil
}

// Mock categories for development/testing when site is not available
func (s *CategoryScraper) CreateMockCategories() error {
	log.Println("Creating mock categories")
	
	// Main categories
	electronics := models.Category{Name: "Electronics", Slug: "electronics", ExternalID: "1001"}
	result := s.db.Where("name = ?", electronics.Name).FirstOrCreate(&electronics)
	if result.Error != nil {
		return result.Error
	}
	
	clothing := models.Category{Name: "Clothing", Slug: "clothing", ExternalID: "1002"}
	result = s.db.Where("name = ?", clothing.Name).FirstOrCreate(&clothing)
	if result.Error != nil {
		return result.Error
	}
	
	homeGarden := models.Category{Name: "Home & Garden", Slug: "home-garden", ExternalID: "1003"}
	result = s.db.Where("name = ?", homeGarden.Name).FirstOrCreate(&homeGarden)
	if result.Error != nil {
		return result.Error
	}
	
	// Sub-categories for Electronics
	smartphones := models.Category{
		Name: "Smartphones", 
		ParentID: &electronics.ID, 
		Slug: "electronics/smartphones", 
		ExternalID: "2001",
	}
	result = s.db.Where("name = ? AND parent_id = ?", smartphones.Name, electronics.ID).FirstOrCreate(&smartphones)
	if result.Error != nil {
		return result.Error
	}
	
	laptops := models.Category{
		Name: "Laptops", 
		ParentID: &electronics.ID, 
		Slug: "electronics/laptops", 
		ExternalID: "2002",
	}
	result = s.db.Where("name = ? AND parent_id = ?", laptops.Name, electronics.ID).FirstOrCreate(&laptops)
	if result.Error != nil {
		return result.Error
	}
	
	// Sub-categories for Clothing
	mensClothing := models.Category{
		Name: "Men's Clothing", 
		ParentID: &clothing.ID, 
		Slug: "clothing/mens", 
		ExternalID: "2003",
	}
	result = s.db.Where("name = ? AND parent_id = ?", mensClothing.Name, clothing.ID).FirstOrCreate(&mensClothing)
	if result.Error != nil {
		return result.Error
	}
	
	womensClothing := models.Category{
		Name: "Women's Clothing", 
		ParentID: &clothing.ID, 
		Slug: "clothing/womens", 
		ExternalID: "2004",
	}
	result = s.db.Where("name = ? AND parent_id = ?", womensClothing.Name, clothing.ID).FirstOrCreate(&womensClothing)
	if result.Error != nil {
		return result.Error
	}
	
	log.Println("Mock categories created successfully")
	return nil
}

// Helper functions
func extractSlug(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func extractCategoryID(url string) string {
	// The logic here would depend on how Trendyol structures their URLs
	// For example, if URLs are like "/category/123-electronics"
	parts := strings.Split(url, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}