package PDFToImage

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"

	"github.com/phpdave11/gofpdi"
	"github.com/signintech/gopdf"
	"gopkg.in/gographics/imagick.v2/imagick"
)

func randomSuffString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return "_" + string(s) + ".jpg"
}

func getCurrentBoxType(box_map map[string]map[string]float64) (string, error) {
	if box_map["/ArtBox"]["w"] != 0 {
		return "/ArtBox", nil
	} else if box_map["/BleedBox"]["w"] != 0 {
		return "/BleedBox", nil
	} else if box_map["/CropBox"]["w"] != 0 {
		return "/CropBox", nil
	} else if box_map["/MediaBox"]["w"] != 0 {
		return "/MediaBox", nil
	} else if box_map["/TrimBox"]["w"] != 0 {
		return "/TrimBox", nil
	} else {
		return "", errors.New("error")
	}
}

// ConvertPdfToJpg will take a filename of a pdf file and convert the file into an
// image which will be saved back to the same location. It will save the image as a
// high resolution jpg file with minimal compression.
func convertPdfToJpg(pdfName string, imageName string, wg_convertPdfToJpg *sync.WaitGroup) error {
	defer wg_convertPdfToJpg.Done()
	defer os.Remove(pdfName)
	// Setup
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Must be *before* ReadImageFile
	// Make sure our image is high quality
	if err := mw.SetResolution(300, 300); err != nil {
		return err
	}

	// Load the image file into imagick
	if err := mw.ReadImage(pdfName); err != nil {
		return err
	}

	// Must be *after* ReadImageFile
	// Flatten image and remove alpha channel, to prevent alpha turning black in jpg
	if err := mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_FLATTEN); err != nil {
		return err
	}

	// Set any compression (100 = max quality)
	if err := mw.SetCompressionQuality(95); err != nil {
		return err
	}

	// Select only first page of pdf
	mw.SetIteratorIndex(0)

	// Convert into JPG
	if err := mw.SetFormat("jpg"); err != nil {
		return err
	}

	// Save File
	return mw.WriteImage(imageName)
}

// Split the PDF into multiple images.
// Incoming parameters: pdf file name, prefix to image.
// Return parameters: if successful, returns an array of converted image file names, otherwise error is returned.
func ConvertPdfToJpg(pdfName string, imageNamePrefix string) ([]string, error) {
	var result []string
	pdf_file := gofpdi.NewImporter()
	pdf_file.SetSourceFile(pdfName)
	page_size := pdf_file.GetNumPages()
	pdf_map_all := pdf_file.GetPageSizes()
	var wg_ConvertPdfToJpg sync.WaitGroup
	wg_ConvertPdfToJpg.Add(page_size)
	for i := 1; i <= page_size; i++ {
		box_type, err := getCurrentBoxType(pdf_map_all[i])
		if err != nil {
			return nil, err
		}
		pdf_map := pdf_map_all[i][box_type]
		w_n := pdf_map["w"]
		h_n := pdf_map["h"]
		urx := pdf_map["urx"]
		ury := pdf_map["ury"]

		pdf := gopdf.GoPdf{}
		pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: w_n, H: h_n}})
		pdf.AddPage()

		// Import page 1
		tpl := pdf.ImportPage(pdfName, i, box_type)

		// Draw pdf onto page
		pdf.UseImportedTemplate(tpl, 0, 0, urx, ury)

		// Draw Image
		suff_string := fmt.Sprint(i) + randomSuffString(10)
		new_image_name := imageNamePrefix + suff_string
		split_pdf := "temp_pdf_page_" + fmt.Sprint(i) + ".pdf"
		err = pdf.WritePdf(split_pdf)
		if err != nil {
			return nil, err
		}
		result = append(result, new_image_name)
		go convertPdfToJpg(split_pdf, new_image_name, &wg_ConvertPdfToJpg)
	}

	wg_ConvertPdfToJpg.Wait()
	return result, nil
}
