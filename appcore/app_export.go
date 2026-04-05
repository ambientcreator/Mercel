package appcore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type actPDFData struct {
	ActNumber        int
	EmployeeFullName string
	ContractNumber   string
	ContractTitle    string
	CustomerName     string
	ContractDate     time.Time
	GeneratedAt      time.Time
	TargetAmount     int
	TotalAmount      int
	Items            []CalculationItem
	TotalWords       string
}

type contractInfo struct {
	Code                  string
	Title                 string
	CustomerName          string
	CustomerDirectorName  string
	CustomerDirectorShort string
}

func (a *App) ExportCurrentCalculationPDF(req ExportCalculationRequest) (string, error) {
	user, err := a.requireAuth()
	if err != nil {
		return "", err
	}

	data, err := buildActPDFData(req, user)
	if err != nil {
		return "", err
	}
	if a.ctx == nil {
		return "", errors.New("\u041a\u043e\u043d\u0442\u0435\u043a\u0441\u0442 \u043f\u0440\u0438\u043b\u043e\u0436\u0435\u043d\u0438\u044f \u0435\u0449\u0451 \u043d\u0435 \u0438\u043d\u0438\u0446\u0438\u0430\u043b\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u043d.")
	}

	filename := fmt.Sprintf("\u0410\u043a\u0442_%d_%s.pdf", data.ActNumber, data.GeneratedAt.Format("02.01.2006"))
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "\u0421\u043e\u0445\u0440\u0430\u043d\u0438\u0442\u044c \u0430\u043a\u0442 \u0432 PDF",
		DefaultFilename: filename,
		Filters: []runtime.FileFilter{{
			DisplayName: "PDF \u0434\u043e\u043a\u0443\u043c\u0435\u043d\u0442 (*.pdf)",
			Pattern:     "*.pdf",
		}},
	})
	if err != nil {
		return "", fmt.Errorf("open save dialog: %w", err)
	}
	if strings.TrimSpace(path) == "" {
		return "", errors.New("\u0421\u043e\u0445\u0440\u0430\u043d\u0435\u043d\u0438\u0435 \u0444\u0430\u0439\u043b\u0430 \u043e\u0442\u043c\u0435\u043d\u0435\u043d\u043e.")
	}
	if err := renderActPDF(path, data); err != nil {
		return "", err
	}
	return path, nil
}

func buildActPDFData(req ExportCalculationRequest, currentUser *User) (actPDFData, error) {
	if req.ActNumber <= 0 {
		return actPDFData{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u043d\u043e\u043c\u0435\u0440 \u0430\u043a\u0442\u0430 \u0431\u043e\u043b\u044c\u0448\u0435 0.")
	}

	employeeName := strings.TrimSpace(req.EmployeeFullName)
	if employeeName == "" && currentUser != nil {
		employeeName = strings.TrimSpace(currentUser.FullName)
	}
	if employeeName == "" {
		return actPDFData{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u0424\u0418\u041e \u0441\u043e\u0442\u0440\u0443\u0434\u043d\u0438\u043a\u0430 \u0434\u043b\u044f \u0432\u044b\u0433\u0440\u0443\u0437\u043a\u0438 \u0430\u043a\u0442\u0430.")
	}

	contractCode := strings.TrimSpace(req.ContractCode)
	if contractCode == "" {
		return actPDFData{}, errors.New("\u0412\u044b\u0431\u0435\u0440\u0438\u0442\u0435 \u0434\u043e\u0433\u043e\u0432\u043e\u0440 \u0434\u043b\u044f \u0430\u043a\u0442\u0430.")
	}
	contract, err := resolveContractInfo(contractCode)
	if err != nil {
		return actPDFData{}, err
	}
	contractNumber := strings.TrimSpace(req.ContractNumber)
	if contractNumber == "" {
		return actPDFData{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u043d\u043e\u043c\u0435\u0440 \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430 \u0434\u043b\u044f \u0432\u044b\u0431\u0440\u0430\u043d\u043d\u043e\u0433\u043e \u0448\u0430\u0431\u043b\u043e\u043d\u0430.")
	}

	contractDateRaw := strings.TrimSpace(req.ContractDate)
	if contractDateRaw == "" {
		return actPDFData{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u0434\u0430\u0442\u0443 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430.")
	}
	contractDate, err := time.Parse("2006-01-02", contractDateRaw)
	if err != nil {
		return actPDFData{}, errors.New("\u041d\u0435 \u0443\u0434\u0430\u043b\u043e\u0441\u044c \u0440\u0430\u0441\u043f\u043e\u0437\u043d\u0430\u0442\u044c \u0434\u0430\u0442\u0443 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430.")
	}

	if len(req.Items) == 0 {
		return actPDFData{}, errors.New("\u0421\u043d\u0430\u0447\u0430\u043b\u0430 \u0432\u044b\u043f\u043e\u043b\u043d\u0438\u0442\u0435 \u0440\u0430\u0441\u0447\u0451\u0442, \u0447\u0442\u043e\u0431\u044b \u0441\u043a\u0430\u0447\u0430\u0442\u044c \u0430\u043a\u0442.")
	}
	if req.TargetAmount <= 0 {
		return actPDFData{}, errors.New("\u0421\u0443\u043c\u043c\u0430 \u0440\u0430\u0441\u0447\u0451\u0442\u0430 \u0434\u043e\u043b\u0436\u043d\u0430 \u0431\u044b\u0442\u044c \u0431\u043e\u043b\u044c\u0448\u0435 0.")
	}

	generatedAt := time.Now()
	if strings.TrimSpace(req.GeneratedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.GeneratedAt)
		if err != nil {
			return actPDFData{}, errors.New("\u041d\u0435 \u0443\u0434\u0430\u043b\u043e\u0441\u044c \u0440\u0430\u0441\u043f\u043e\u0437\u043d\u0430\u0442\u044c \u0434\u0430\u0442\u0443 \u0440\u0430\u0441\u0447\u0451\u0442\u0430 \u0434\u043b\u044f \u0430\u043a\u0442\u0430.")
		}
		generatedAt = parsed
	}

	total := 0
	for _, item := range req.Items {
		total += item.LineTotal
	}
	if total != req.TargetAmount {
		return actPDFData{}, fmt.Errorf("\u042d\u043a\u0441\u043f\u043e\u0440\u0442 \u0434\u043e\u0441\u0442\u0443\u043f\u0435\u043d \u0442\u043e\u043b\u044c\u043a\u043e \u0434\u043b\u044f \u0442\u043e\u0447\u043d\u043e\u0433\u043e \u0440\u0430\u0441\u0447\u0451\u0442\u0430 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443 %s.", displayMoney(req.TargetAmount))
	}

	totalWords, err := numberToRussianWords(total)
	if err != nil {
		return actPDFData{}, err
	}

	return actPDFData{
		ActNumber:        req.ActNumber,
		EmployeeFullName: employeeName,
		ContractNumber:   contractNumber,
		ContractTitle:    contract.Title,
		CustomerName:     contract.CustomerName,
		ContractDate:     contractDate,
		GeneratedAt:      generatedAt,
		TargetAmount:     req.TargetAmount,
		TotalAmount:      total,
		Items:            req.Items,
		TotalWords:       totalWords,
	}, nil
}

func resolveContractInfo(code string) (contractInfo, error) {
	switch strings.TrimSpace(code) {
	case "1":
		return contractInfo{
			Code:                  "1",
			Title:                 "\u00ab\u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433\u0441\u043a\u0438\u0435 \u043a\u043e\u043c\u043f\u044c\u044e\u0442\u0435\u0440\u043d\u044b\u0435 \u0441\u0435\u0442\u0438\u00bb \u21161",
			CustomerName:          "\u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433\u0441\u043a\u0438\u0435 \u043a\u043e\u043c\u043f\u044c\u044e\u0442\u0435\u0440\u043d\u044b\u0435 \u0441\u0435\u0442\u0438",
			CustomerDirectorName:  "\u0422\u0438\u0442\u043e\u0432\u0430 \u0418\u0433\u043e\u0440\u044f \u0421\u0432\u0435\u0442\u043e\u0437\u0430\u0440\u043e\u0432\u0438\u0447\u0430",
			CustomerDirectorShort: "\u0422\u0438\u0442\u043e\u0432 \u0418.\u0421.",
		}, nil
	case "2":
		return contractInfo{
			Code:                  "2",
			Title:                 "\u00ab\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c\u00bb \u21162",
			CustomerName:          "\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c",
			CustomerDirectorName:  "\u0422\u0438\u0442\u043e\u0432\u0430 \u0418\u0433\u043e\u0440\u044f \u0421\u0432\u0435\u0442\u043e\u0437\u0430\u0440\u043e\u0432\u0438\u0447\u0430",
			CustomerDirectorShort: "\u0422\u0438\u0442\u043e\u0432 \u0418.\u0421.",
		}, nil
	default:
		return contractInfo{}, errors.New("\u0412\u044b\u0431\u0435\u0440\u0438\u0442\u0435 \u0434\u043e\u0433\u043e\u0432\u043e\u0440 \u0434\u043b\u044f \u0430\u043a\u0442\u0430: \u00ab\u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433\u0441\u043a\u0438\u0435 \u043a\u043e\u043c\u043f\u044c\u044e\u0442\u0435\u0440\u043d\u044b\u0435 \u0441\u0435\u0442\u0438\u00bb \u21161 \u0438\u043b\u0438 \u00ab\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c\u00bb \u21162.")
	}
}

func renderActPDF(path string, data actPDFData) error {
	fontRegular, fontBold, fontItalic, err := resolvePDFFonts()
	if err != nil {
		return err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(14, 16, 14)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddUTF8Font("Mercel", "", fontRegular)
	pdf.AddUTF8Font("Mercel", "B", fontBold)
	pdf.AddUTF8Font("Mercel", "I", fontItalic)
	pdf.SetFont("Mercel", "", 10)
	pdf.AddPage()

	pageBottom := 287.0
	if err := ensureActFitsOnePage(pdf, data, pageBottom); err != nil {
		return err
	}

	drawActPDF(pdf, data)
	if pdf.PageNo() > 1 {
		return errors.New("\u0414\u043e\u043a\u0443\u043c\u0435\u043d\u0442 \u043d\u0435 \u043f\u043e\u043c\u0435\u0449\u0430\u0435\u0442\u0441\u044f \u043d\u0430 \u043e\u0434\u0438\u043d \u043b\u0438\u0441\u0442 PDF. \u0423\u043c\u0435\u043d\u044c\u0448\u0438\u0442\u0435 \u043a\u043e\u043b\u0438\u0447\u0435\u0441\u0442\u0432\u043e \u0443\u0441\u043b\u0443\u0433 \u0438\u043b\u0438 \u0434\u043b\u0438\u043d\u0443 \u043d\u0430\u0437\u0432\u0430\u043d\u0438\u0439.")
	}
	if err := pdf.OutputFileAndClose(path); err != nil {
		return fmt.Errorf("save pdf: %w", err)
	}
	return nil
}

func resolvePDFFonts() (string, string, string, error) {
	candidates := [][3]string{
		{`C:\Windows\Fonts\arial.ttf`, `C:\Windows\Fonts\arialbd.ttf`, `C:\Windows\Fonts\ariali.ttf`},
		{`C:\Windows\Fonts\segoeui.ttf`, `C:\Windows\Fonts\segoeuib.ttf`, `C:\Windows\Fonts\segoeuii.ttf`},
	}
	for _, candidate := range candidates {
		if fileExists(candidate[0]) && fileExists(candidate[1]) && fileExists(candidate[2]) {
			return candidate[0], candidate[1], candidate[2], nil
		}
	}
	if windir := os.Getenv("WINDIR"); windir != "" {
		regular := filepath.Join(windir, "Fonts", "arial.ttf")
		bold := filepath.Join(windir, "Fonts", "arialbd.ttf")
		italic := filepath.Join(windir, "Fonts", "ariali.ttf")
		if fileExists(regular) && fileExists(bold) && fileExists(italic) {
			return regular, bold, italic, nil
		}
	}
	return "", "", "", errors.New("\u041d\u0435 \u0443\u0434\u0430\u043b\u043e\u0441\u044c \u043d\u0430\u0439\u0442\u0438 \u0441\u0438\u0441\u0442\u0435\u043c\u043d\u044b\u0435 \u0448\u0440\u0438\u0444\u0442\u044b \u0434\u043b\u044f \u0433\u0435\u043d\u0435\u0440\u0430\u0446\u0438\u0438 PDF. \u0423\u0431\u0435\u0434\u0438\u0442\u0435\u0441\u044c, \u0447\u0442\u043e \u0432 Windows \u0434\u043e\u0441\u0442\u0443\u043f\u043d\u044b Arial \u0438\u043b\u0438 Segoe UI.")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ensureActFitsOnePage(pdf *gofpdf.Fpdf, data actPDFData, pageBottom float64) error {
	pdf.SetFont("Mercel", "", 10)
	rowHeights := actRowHeights(pdf, data.Items, 106, 4.9)
	totalRowsHeight := 0.0
	for _, height := range rowHeights {
		totalRowsHeight += height
	}

	estimatedHeight := 16.0
	estimatedHeight += 11.0
	estimatedHeight += 10.0
	estimatedHeight += 28.0
	estimatedHeight += 8.0
	estimatedHeight += totalRowsHeight
	estimatedHeight += 11.0
	estimatedHeight += 12.0
	estimatedHeight += 18.0
	estimatedHeight += 24.0

	if estimatedHeight > pageBottom {
		return errors.New("\u0414\u043e\u043a\u0443\u043c\u0435\u043d\u0442 \u043d\u0435 \u043f\u043e\u043c\u0435\u0449\u0430\u0435\u0442\u0441\u044f \u043d\u0430 \u043e\u0434\u0438\u043d \u043b\u0438\u0441\u0442 PDF. \u0423\u043c\u0435\u043d\u044c\u0448\u0438\u0442\u0435 \u043a\u043e\u043b\u0438\u0447\u0435\u0441\u0442\u0432\u043e \u0443\u0441\u043b\u0443\u0433 \u0438\u043b\u0438 \u0434\u043b\u0438\u043d\u0443 \u043d\u0430\u0437\u0432\u0430\u043d\u0438\u0439.")
	}
	return nil
}

func drawActPDF(pdf *gofpdf.Fpdf, data actPDFData) {
	left := 14.0
	top := 18.0
	pageWidth, _ := pdf.GetPageSize()
	contentWidth := pageWidth - 28.0

	pdf.SetXY(left, top)
	pdf.SetFont("Mercel", "B", 10)
	pdf.CellFormat(contentWidth, 6, fmt.Sprintf("\u0410\u041a\u0422 \u2116 %d \u041f\u0420\u0418\u0401\u041c\u0410-\u0421\u0414\u0410\u0427\u0418 \u041e\u041a\u0410\u0417\u0410\u041d\u041d\u042b\u0425 \u0423\u0421\u041b\u0423\u0413", data.ActNumber), "", 1, "C", false, 0, "")

	pdf.SetFont("Mercel", "", 9)
	pdf.SetX(left)
	pdf.CellFormat(contentWidth/2, 6, "\u0433. \u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433", "", 0, "L", false, 0, "")
	pdf.CellFormat(contentWidth/2, 6, russianActDate(data.GeneratedAt), "", 1, "R", false, 0, "")

	pdf.Ln(7)
	pdf.SetX(left)
	pdf.SetFont("Mercel", "", 8.8)
	intro := fmt.Sprintf(
		"\u041e\u0431\u0449\u0435\u0441\u0442\u0432\u043e \u0441 \u043e\u0433\u0440\u0430\u043d\u0438\u0447\u0435\u043d\u043d\u043e\u0439 \u043e\u0442\u0432\u0435\u0442\u0441\u0442\u0432\u0435\u043d\u043d\u043e\u0441\u0442\u044c\u044e \u00ab%s\u00bb, \u0438\u043c\u0435\u043d\u0443\u0435\u043c\u043e\u0435 \u0432 \u0434\u0430\u043b\u044c\u043d\u0435\u0439\u0448\u0435\u043c \u0417\u0430\u043a\u0430\u0437\u0447\u0438\u043a, \u0432 \u043b\u0438\u0446\u0435 \u0433\u0435\u043d\u0435\u0440\u0430\u043b\u044c\u043d\u043e\u0433\u043e \u0434\u0438\u0440\u0435\u043a\u0442\u043e\u0440\u0430 %s \u0441 \u043e\u0434\u043d\u043e\u0439 \u0441\u0442\u043e\u0440\u043e\u043d\u044b \u0438 \u0418\u043d\u0434\u0438\u0432\u0438\u0434\u0443\u0430\u043b\u044c\u043d\u044b\u0439 \u043f\u0440\u0435\u0434\u043f\u0440\u0438\u043d\u0438\u043c\u0430\u0442\u0435\u043b\u044c %s, \u0438\u043c\u0435\u043d\u0443\u0435\u043c\u044b\u0439 \u0432 \u0434\u0430\u043b\u044c\u043d\u0435\u0439\u0448\u0435\u043c \u0418\u0441\u043f\u043e\u043b\u043d\u0438\u0442\u0435\u043b\u044c, \u0441 \u0434\u0440\u0443\u0433\u043e\u0439 \u0441\u0442\u043e\u0440\u043e\u043d\u044b \u0441\u043e\u0441\u0442\u0430\u0432\u0438\u043b\u0438 \u043d\u0430\u0441\u0442\u043e\u044f\u0449\u0438\u0439 \u0430\u043a\u0442, \u0441\u043e\u0433\u043b\u0430\u0441\u043d\u043e \u0414\u043e\u0433\u043e\u0432\u043e\u0440\u0443 %s \u043e\u0442 %s, \u0437\u0430\u043a\u043b\u044e\u0447\u0435\u043d\u043d\u043e\u0433\u043e \u043c\u0435\u0436\u0434\u0443 \u0421\u0442\u043e\u0440\u043e\u043d\u0430\u043c\u0438, \u043e \u0442\u043e\u043c, \u0447\u0442\u043e \u0418\u0441\u043f\u043e\u043b\u043d\u0438\u0442\u0435\u043b\u044c \u0432\u044b\u043f\u043e\u043b\u043d\u0438\u043b, \u0430 \u0417\u0430\u043a\u0430\u0437\u0447\u0438\u043a \u043f\u0440\u0438\u043d\u044f\u043b \u0441\u043b\u0435\u0434\u0443\u044e\u0449\u0438\u0435 \u0440\u0430\u0431\u043e\u0442\u044b:",
		data.CustomerName,
		resolveContractDirectorName(data.ContractTitle),
		data.EmployeeFullName,
		data.ContractTitle,
		formatContractDateLong(data.ContractDate),
	)
	pdf.MultiCell(contentWidth, 4.7, intro, "", "J", false)
	pdf.Ln(2)

	drawActTable(pdf, left, data)

	pdf.Ln(4)
	pdf.SetX(left)
	pdf.SetFont("Mercel", "I", 8)
	pdf.MultiCell(contentWidth, 4.2, fmt.Sprintf("\u0412\u0441\u0435\u0433\u043e \u0432\u044b\u043f\u043e\u043b\u043d\u0435\u043d\u043e \u0440\u0430\u0431\u043e\u0442 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443: %s \u0440\u0443\u0431\u043b\u0435\u0439 00 \u043a\u043e\u043f\u0435\u0435\u043a \u0431\u0435\u0437 \u041d\u0414\u0421.", data.TotalWords), "", "L", false)

	pdf.Ln(5)
	pdf.SetX(left)
	pdf.SetFont("Mercel", "", 8.6)
	pdf.MultiCell(contentWidth, 4.8, "\u0412\u044b\u0448\u0435\u043f\u0435\u0440\u0435\u0447\u0438\u0441\u043b\u0435\u043d\u043d\u044b\u0435 \u0443\u0441\u043b\u0443\u0433\u0438 \u0432\u044b\u043f\u043e\u043b\u043d\u0435\u043d\u044b \u043f\u043e\u043b\u043d\u043e\u0441\u0442\u044c\u044e \u0438 \u0432 \u0441\u0440\u043e\u043a. \u0417\u0430\u043a\u0430\u0437\u0447\u0438\u043a \u043f\u0440\u0435\u0442\u0435\u043d\u0437\u0438\u0439 \u043f\u043e \u043e\u0431\u044a\u0451\u043c\u0443, \u043a\u0430\u0447\u0435\u0441\u0442\u0432\u0443 \u0438 \u0441\u0440\u043e\u043a\u0430\u043c \u043e\u043a\u0430\u0437\u0430\u043d\u0438\u044f \u0443\u0441\u043b\u0443\u0433 \u043d\u0435 \u0438\u043c\u0435\u0435\u0442.", "", "J", false)

	pdf.Ln(6)
	signatureY := pdf.GetY()
	colWidth := contentWidth / 2

	pdf.SetFont("Mercel", "B", 9)
	pdf.SetXY(left, signatureY)
	pdf.CellFormat(colWidth, 5, "\u0418\u0441\u043f\u043e\u043b\u043d\u0438\u0442\u0435\u043b\u044c:", "", 0, "L", false, 0, "")
	pdf.CellFormat(colWidth, 5, "\u0417\u0430\u043a\u0430\u0437\u0447\u0438\u043a:", "", 1, "L", false, 0, "")

	pdf.SetFont("Mercel", "", 8.8)
	pdf.SetX(left)
	drawSignature(pdf, colWidth, shortEmployeeSignatureName(data.EmployeeFullName))
	drawSignature(pdf, colWidth, resolveContractDirectorShort(data.ContractTitle))
}

func resolveContractDirectorName(contractTitle string) string {
	info, err := resolveContractInfoFromTitle(contractTitle)
	if err != nil {
		return "\u0422\u0438\u0442\u043e\u0432\u0430 \u0418\u0433\u043e\u0440\u044f \u0421\u0432\u0435\u0442\u043e\u0437\u0430\u0440\u043e\u0432\u0438\u0447\u0430"
	}
	return info.CustomerDirectorName
}

func resolveContractDirectorShort(contractTitle string) string {
	info, err := resolveContractInfoFromTitle(contractTitle)
	if err != nil {
		return "\u0422\u0438\u0442\u043e\u0432 \u0418.\u0421."
	}
	return info.CustomerDirectorShort
}

func resolveContractInfoFromTitle(title string) (contractInfo, error) {
	for _, code := range []string{"1", "2"} {
		info, err := resolveContractInfo(code)
		if err == nil && info.Title == title {
			return info, nil
		}
	}
	return contractInfo{}, errors.New("\u043d\u0435\u0438\u0437\u0432\u0435\u0441\u0442\u043d\u044b\u0439 \u0448\u0430\u0431\u043b\u043e\u043d \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430")
}

func drawSignature(pdf *gofpdf.Fpdf, width float64, name string) {
	nameWidth := width - 28
	if nameWidth < 60 {
		nameWidth = width * 0.7
	}
	pdf.CellFormat(nameWidth, 16, name, "", 0, "L", false, 0, "")
	pdf.CellFormat(width-nameWidth, 16, "   /________/", "", 0, "L", false, 0, "")
}

func drawActTable(pdf *gofpdf.Fpdf, left float64, data actPDFData) {
	colNo := 9.0
	colName := 106.0
	colQty := 28.0
	colMoney := 33.0
	lineHeight := 4.9
	paddingTop := 2.1
	paddingHorizontal := 2.0

	pdf.SetFont("Mercel", "B", 8.8)
	pdf.SetX(left)
	pdf.CellFormat(colNo, 8, "\u2116", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colName, 8, "\u041d\u0430\u0438\u043c\u0435\u043d\u043e\u0432\u0430\u043d\u0438\u0435 \u0443\u0441\u043b\u0443\u0433\u0438", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colQty, 8, "\u041a\u043e\u043b\u0438\u0447\u0435\u0441\u0442\u0432\u043e", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colMoney, 8, "\u0421\u0442\u043e\u0438\u043c\u043e\u0441\u0442\u044c", "1", 1, "C", false, 0, "")

	pdf.SetFont("Mercel", "", 8.4)
	for index, item := range data.Items {
		lines := pdf.SplitText(item.Name, colName-(paddingHorizontal*2))
		if len(lines) == 0 {
			lines = []string{""}
		}
		rowHeight := float64(len(lines))*lineHeight + (paddingTop * 2)
		if rowHeight < 10 {
			rowHeight = 10
		}

		x := left
		y := pdf.GetY()
		pdf.Rect(x, y, colNo, rowHeight, "")
		pdf.Rect(x+colNo, y, colName, rowHeight, "")
		pdf.Rect(x+colNo+colName, y, colQty, rowHeight, "")
		pdf.Rect(x+colNo+colName+colQty, y, colMoney, rowHeight, "")

		pdf.SetXY(x, y)
		pdf.CellFormat(colNo, rowHeight, fmt.Sprintf("%d", index+1), "", 0, "C", false, 0, "")

		pdf.SetXY(x+colNo+paddingHorizontal, y+paddingTop)
		pdf.MultiCell(colName-(paddingHorizontal*2), lineHeight, item.Name, "", "L", false)

		pdf.SetXY(x+colNo+colName, y)
		pdf.CellFormat(colQty, rowHeight, formatDocumentQuantity(item.Quantity, item.Unit), "", 0, "C", false, 0, "")

		pdf.SetXY(x+colNo+colName+colQty, y)
		pdf.CellFormat(colMoney, rowHeight, formatDocumentMoney(item.LineTotal), "", 1, "C", false, 0, "")
	}

	pdf.SetFont("Mercel", "B", 9)
	pdf.SetX(left + colNo + colName)
	pdf.CellFormat(colQty, 10, "\u0418\u0422\u041e\u0413\u041e:", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colMoney, 10, formatDocumentMoney(data.TotalAmount), "1", 1, "C", false, 0, "")
}

func actRowHeights(pdf *gofpdf.Fpdf, items []CalculationItem, nameWidth float64, lineHeight float64) []float64 {
	heights := make([]float64, 0, len(items))
	for _, item := range items {
		lines := pdf.SplitText(item.Name, nameWidth-4)
		height := float64(len(lines))*lineHeight + 4.2
		if height < 10 {
			height = 10
		}
		heights = append(heights, height)
	}
	return heights
}

func russianActDate(value time.Time) string {
	month := russianMonthGenitive(int(value.Month()))
	return fmt.Sprintf("\u00ab%02d\u00bb %s %d \u0433.", value.Day(), month, value.Year())
}

func formatContractDateLong(value time.Time) string {
	month := russianMonthGenitive(int(value.Month()))
	return fmt.Sprintf("\u00ab%02d\u00bb %s %d \u0433\u043e\u0434\u0430", value.Day(), month, value.Year())
}

func russianMonthGenitive(month int) string {
	months := []string{
		"\u044f\u043d\u0432\u0430\u0440\u044f", "\u0444\u0435\u0432\u0440\u0430\u043b\u044f", "\u043c\u0430\u0440\u0442\u0430", "\u0430\u043f\u0440\u0435\u043b\u044f", "\u043c\u0430\u044f", "\u0438\u044e\u043d\u044f",
		"\u0438\u044e\u043b\u044f", "\u0430\u0432\u0433\u0443\u0441\u0442\u0430", "\u0441\u0435\u043d\u0442\u044f\u0431\u0440\u044f", "\u043e\u043a\u0442\u044f\u0431\u0440\u044f", "\u043d\u043e\u044f\u0431\u0440\u044f", "\u0434\u0435\u043a\u0430\u0431\u0440\u044f",
	}
	if month < 1 || month > len(months) {
		return ""
	}
	return months[month-1]
}

func formatDocumentMoney(value int) string {
	return fmt.Sprintf("%s \u0440\u0443\u0431.", formatThousands(value))
}

func formatDocumentQuantity(quantity int, unit string) string {
	return fmt.Sprintf("%d %s", quantity, normalizeUnit(unit))
}

func formatThousands(value int) string {
	source := fmt.Sprintf("%d", value)
	if len(source) <= 3 {
		return source
	}
	parts := make([]string, 0, (len(source)+2)/3)
	for len(source) > 3 {
		parts = append([]string{source[len(source)-3:]}, parts...)
		source = source[:len(source)-3]
	}
	if source != "" {
		parts = append([]string{source}, parts...)
	}
	return strings.Join(parts, " ")
}

func shortEmployeeSignatureName(fullName string) string {
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	last := parts[0]
	initials := make([]string, 0, 2)
	for _, part := range parts[1:] {
		runes := []rune(part)
		if len(runes) > 0 {
			initials = append(initials, string(runes[0])+".")
		}
	}
	return strings.TrimSpace(last + " " + strings.Join(initials, ""))
}

func numberToRussianWords(n int) (string, error) {
	if n < 0 {
		return "", errors.New("\u0421\u0443\u043c\u043c\u0430 \u0434\u043b\u044f \u043f\u0440\u043e\u043f\u0438\u0441\u0438 \u043d\u0435 \u043c\u043e\u0436\u0435\u0442 \u0431\u044b\u0442\u044c \u043e\u0442\u0440\u0438\u0446\u0430\u0442\u0435\u043b\u044c\u043d\u043e\u0439.")
	}
	if n == 0 {
		return "\u043d\u043e\u043b\u044c", nil
	}
	if n > 999999999 {
		return "", errors.New("\u0421\u0443\u043c\u043c\u0430 \u0434\u043b\u044f \u043f\u0440\u043e\u043f\u0438\u0441\u0438 \u0441\u043b\u0438\u0448\u043a\u043e\u043c \u0431\u043e\u043b\u044c\u0448\u0430\u044f.")
	}

	unitsMale := []string{"", "\u043e\u0434\u0438\u043d", "\u0434\u0432\u0430", "\u0442\u0440\u0438", "\u0447\u0435\u0442\u044b\u0440\u0435", "\u043f\u044f\u0442\u044c", "\u0448\u0435\u0441\u0442\u044c", "\u0441\u0435\u043c\u044c", "\u0432\u043e\u0441\u0435\u043c\u044c", "\u0434\u0435\u0432\u044f\u0442\u044c"}
	unitsFemale := []string{"", "\u043e\u0434\u043d\u0430", "\u0434\u0432\u0435", "\u0442\u0440\u0438", "\u0447\u0435\u0442\u044b\u0440\u0435", "\u043f\u044f\u0442\u044c", "\u0448\u0435\u0441\u0442\u044c", "\u0441\u0435\u043c\u044c", "\u0432\u043e\u0441\u0435\u043c\u044c", "\u0434\u0435\u0432\u044f\u0442\u044c"}
	teens := []string{"\u0434\u0435\u0441\u044f\u0442\u044c", "\u043e\u0434\u0438\u043d\u043d\u0430\u0434\u0446\u0430\u0442\u044c", "\u0434\u0432\u0435\u043d\u0430\u0434\u0446\u0430\u0442\u044c", "\u0442\u0440\u0438\u043d\u0430\u0434\u0446\u0430\u0442\u044c", "\u0447\u0435\u0442\u044b\u0440\u043d\u0430\u0434\u0446\u0430\u0442\u044c", "\u043f\u044f\u0442\u043d\u0430\u0434\u0446\u0430\u0442\u044c", "\u0448\u0435\u0441\u0442\u043d\u0430\u0434\u0446\u0430\u0442\u044c", "\u0441\u0435\u043c\u043d\u0430\u0434\u0446\u0430\u0442\u044c", "\u0432\u043e\u0441\u0435\u043c\u043d\u0430\u0434\u0446\u0430\u0442\u044c", "\u0434\u0435\u0432\u044f\u0442\u043d\u0430\u0434\u0446\u0430\u0442\u044c"}
	tens := []string{"", "", "\u0434\u0432\u0430\u0434\u0446\u0430\u0442\u044c", "\u0442\u0440\u0438\u0434\u0446\u0430\u0442\u044c", "\u0441\u043e\u0440\u043e\u043a", "\u043f\u044f\u0442\u044c\u0434\u0435\u0441\u044f\u0442", "\u0448\u0435\u0441\u0442\u044c\u0434\u0435\u0441\u044f\u0442", "\u0441\u0435\u043c\u044c\u0434\u0435\u0441\u044f\u0442", "\u0432\u043e\u0441\u0435\u043c\u044c\u0434\u0435\u0441\u044f\u0442", "\u0434\u0435\u0432\u044f\u043d\u043e\u0441\u0442\u043e"}
	hundreds := []string{"", "\u0441\u0442\u043e", "\u0434\u0432\u0435\u0441\u0442\u0438", "\u0442\u0440\u0438\u0441\u0442\u0430", "\u0447\u0435\u0442\u044b\u0440\u0435\u0441\u0442\u0430", "\u043f\u044f\u0442\u044c\u0441\u043e\u0442", "\u0448\u0435\u0441\u0442\u044c\u0441\u043e\u0442", "\u0441\u0435\u043c\u044c\u0441\u043e\u0442", "\u0432\u043e\u0441\u0435\u043c\u044c\u0441\u043e\u0442", "\u0434\u0435\u0432\u044f\u0442\u044c\u0441\u043e\u0442"}

	type groupInfo struct {
		value  int
		forms  [3]string
		female bool
	}

	groups := []groupInfo{
		{value: n / 1000000, forms: [3]string{"\u043c\u0438\u043b\u043b\u0438\u043e\u043d", "\u043c\u0438\u043b\u043b\u0438\u043e\u043d\u0430", "\u043c\u0438\u043b\u043b\u0438\u043e\u043d\u043e\u0432"}},
		{value: (n / 1000) % 1000, forms: [3]string{"\u0442\u044b\u0441\u044f\u0447\u0430", "\u0442\u044b\u0441\u044f\u0447\u0438", "\u0442\u044b\u0441\u044f\u0447"}, female: true},
		{value: n % 1000},
	}

	parts := make([]string, 0, 8)
	for _, group := range groups {
		if group.value == 0 {
			continue
		}
		h := group.value / 100
		t := (group.value / 10) % 10
		u := group.value % 10
		if h > 0 {
			parts = append(parts, hundreds[h])
		}
		if t == 1 {
			parts = append(parts, teens[u])
		} else {
			if t > 1 {
				parts = append(parts, tens[t])
			}
			if u > 0 {
				if group.female {
					parts = append(parts, unitsFemale[u])
				} else {
					parts = append(parts, unitsMale[u])
				}
			}
		}
		if group.forms[0] != "" {
			parts = append(parts, chooseRussianPlural(group.value, group.forms[0], group.forms[1], group.forms[2]))
		}
	}

	return strings.Join(parts, " "), nil
}

func chooseRussianPlural(value int, one string, two string, many string) string {
	value = value % 100
	if value >= 11 && value <= 14 {
		return many
	}
	switch value % 10 {
	case 1:
		return one
	case 2, 3, 4:
		return two
	default:
		return many
	}
}
