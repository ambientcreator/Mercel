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
	TotalCurrency    string
	ActTemplate      string
}

const (
	actPDFTargetScale = 1.30
	actPDFMinScale    = 1.00
)

type actPDFLayout struct {
	Scale                 float64
	Left                  float64
	Right                 float64
	Top                   float64
	Bottom                float64
	ContentWidth          float64
	TitleFont             float64
	MetaFont              float64
	BodyFont              float64
	TotalFont             float64
	ClosingFont           float64
	SignatureFont         float64
	TableHeaderFont       float64
	TableBodyFont         float64
	TableTotalFont        float64
	TitleHeight           float64
	MetaHeight            float64
	IntroLineHeight       float64
	TotalLineHeight       float64
	ClosingLineHeight     float64
	SignatureHeaderHeight float64
	SignatureHeight       float64
	IntroGap              float64
	TableGap              float64
	TotalGap              float64
	ClosingGap            float64
	SignatureGap          float64
	TableHeaderHeight     float64
	TableTotalHeight      float64
	TableLineHeight       float64
	TablePaddingTop       float64
	TablePaddingX         float64
	ColNo                 float64
	ColName               float64
	ColQty                float64
	ColMoney              float64
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

	template, err := a.ensureUserActTemplate(user.ID, user.ActTemplate)
	if err != nil {
		return "", err
	}
	user.ActTemplate = template

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
		return actPDFData{}, errors.New("\u041d\u043e\u043c\u0435\u0440 \u0430\u043a\u0442\u0430 \u0434\u043e\u043b\u0436\u0435\u043d \u0431\u044b\u0442\u044c \u0431\u043e\u043b\u044c\u0448\u0435 0.")
	}

	employeeName := strings.TrimSpace(req.EmployeeFullName)
	if employeeName == "" && currentUser != nil {
		employeeName = strings.TrimSpace(currentUser.FullName)
	}
	if employeeName == "" {
		return actPDFData{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u0424\u0418\u041e \u0441\u043e\u0442\u0440\u0443\u0434\u043d\u0438\u043a\u0430 \u0434\u043b\u044f \u0430\u043a\u0442\u0430.")
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
		return actPDFData{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u043d\u043e\u043c\u0435\u0440 \u0432\u044b\u0431\u0440\u0430\u043d\u043d\u043e\u0433\u043e \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430.")
	}

	contractDateRaw := strings.TrimSpace(req.ContractDate)
	if contractDateRaw == "" {
		return actPDFData{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u0434\u0430\u0442\u0443 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430.")
	}
	contractDate, err := time.Parse("2006-01-02", contractDateRaw)
	if err != nil {
		return actPDFData{}, errors.New("\u0414\u0430\u0442\u0430 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430 \u0443\u043a\u0430\u0437\u0430\u043d\u0430 \u0432 \u043d\u0435\u0432\u0435\u0440\u043d\u043e\u043c \u0444\u043e\u0440\u043c\u0430\u0442\u0435.")
	}

	if len(req.Items) == 0 {
		return actPDFData{}, errors.New("\u041d\u0435 \u043f\u0435\u0440\u0435\u0434\u0430\u043d\u044b \u043f\u043e\u0437\u0438\u0446\u0438\u0438 \u0434\u043b\u044f \u0432\u044b\u0433\u0440\u0443\u0437\u043a\u0438 \u0430\u043a\u0442\u0430.")
	}
	if req.TargetAmount <= 0 {
		return actPDFData{}, errors.New("\u0421\u0443\u043c\u043c\u0430 \u0430\u043a\u0442\u0430 \u0434\u043e\u043b\u0436\u043d\u0430 \u0431\u044b\u0442\u044c \u0431\u043e\u043b\u044c\u0448\u0435 0.")
	}

	generatedAt := time.Now()
	if strings.TrimSpace(req.GeneratedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.GeneratedAt)
		if err != nil {
			return actPDFData{}, errors.New("\u0414\u0430\u0442\u0430 \u0444\u043e\u0440\u043c\u0438\u0440\u043e\u0432\u0430\u043d\u0438\u044f \u0430\u043a\u0442\u0430 \u0443\u043a\u0430\u0437\u0430\u043d\u0430 \u0432 \u043d\u0435\u0432\u0435\u0440\u043d\u043e\u043c \u0444\u043e\u0440\u043c\u0430\u0442\u0435.")
		}
		generatedAt = parsed
	}

	exportItems := make([]CalculationItem, 0, len(req.Items))
	total := 0
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			continue
		}
		if item.AllocationPercent != nil && *item.AllocationPercent <= 0 {
			continue
		}
		exportItems = append(exportItems, item)
		total += item.LineTotal
	}
	if len(exportItems) == 0 {
		return actPDFData{}, errors.New("\u041f\u043e\u0441\u043b\u0435 \u0438\u0441\u043a\u043b\u044e\u0447\u0435\u043d\u0438\u044f \u0443\u0441\u043b\u0443\u0433 \u0441 0% \u0432 \u0430\u043a\u0442\u0435 \u043d\u0435 \u043e\u0441\u0442\u0430\u043b\u043e\u0441\u044c \u043f\u043e\u0437\u0438\u0446\u0438\u0439 \u0434\u043b\u044f \u0432\u044b\u0433\u0440\u0443\u0437\u043a\u0438.")
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
		Items:            exportItems,
		TotalWords:       totalWords,
		TotalCurrency:    rubleNoun(total),
		ActTemplate:      actTemplateOrDefault(currentUserActTemplate(currentUser)),
	}, nil
}

// EN: Function `currentUserActTemplate`.
//
// EN: What it does: currentUserActTemplate reads the blank assigned to the acting user without panicking on a nil session.
//
// EN: Key points: buildActPDFData is also called from tests with a bare user, so a missing blank must simply fall
// EN: back to the classic act instead of failing the export.
func currentUserActTemplate(currentUser *User) string {
	if currentUser == nil {
		return ""
	}
	return currentUser.ActTemplate
}

func resolveContractInfo(code string) (contractInfo, error) {
	switch strings.TrimSpace(code) {
	case "1":
		return contractInfo{
			Code:                  "1",
			Title:                 "\u00ab\u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433\u0441\u043a\u0438\u0435 \u043a\u043e\u043c\u043f\u044c\u044e\u0442\u0435\u0440\u043d\u044b\u0435 \u0441\u0435\u0442\u0438\u00bb",
			CustomerName:          "\u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433\u0441\u043a\u0438\u0435 \u043a\u043e\u043c\u043f\u044c\u044e\u0442\u0435\u0440\u043d\u044b\u0435 \u0441\u0435\u0442\u0438",
			CustomerDirectorName:  "\u0422\u0438\u0442\u043e\u0432\u0430 \u0418\u0433\u043e\u0440\u044f \u0421\u0432\u0435\u0442\u043e\u0437\u0430\u0440\u043e\u0432\u0438\u0447\u0430",
			CustomerDirectorShort: "\u0422\u0438\u0442\u043e\u0432 \u0418.\u0421.",
		}, nil
	case "2":
		return contractInfo{
			Code:                  "2",
			Title:                 "\u00ab\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c\u00bb",
			CustomerName:          "\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c",
			CustomerDirectorName:  "\u0422\u0438\u043c\u043e\u0445\u0438\u043d\u0430 \u041f\u0430\u0432\u043b\u0430 \u0410\u043b\u0435\u043a\u0441\u0430\u043d\u0434\u0440\u043e\u0432\u0438\u0447\u0430",
			CustomerDirectorShort: "\u0422\u0438\u043c\u043e\u0445\u0438\u043d \u041f.\u0410.",
		}, nil
	default:
		return contractInfo{}, errors.New("\u0412\u044b\u0431\u0435\u0440\u0438\u0442\u0435 \u0434\u043e\u0433\u043e\u0432\u043e\u0440 \u0434\u043b\u044f \u0430\u043a\u0442\u0430: \u00ab\u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433\u0441\u043a\u0438\u0435 \u043a\u043e\u043c\u043f\u044c\u044e\u0442\u0435\u0440\u043d\u044b\u0435 \u0441\u0435\u0442\u0438\u00bb \u0438\u043b\u0438 \u00ab\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c\u00bb.")
	}
}

// EN: Function `errActDoesNotFit`.
//
// EN: What it does: errActDoesNotFit is the single message shown when no blank scale keeps the act on one sheet.
//
// EN: Key points: every blank must fit exactly one A4 page, so all renderers report the same actionable hint.
func errActDoesNotFit() error {
	return errors.New("\u0414\u043e\u043a\u0443\u043c\u0435\u043d\u0442 \u043d\u0435 \u043f\u043e\u043c\u0435\u0449\u0430\u0435\u0442\u0441\u044f \u043d\u0430 \u043e\u0434\u0438\u043d \u043b\u0438\u0441\u0442 PDF. \u0423\u043c\u0435\u043d\u044c\u0448\u0438\u0442\u0435 \u043a\u043e\u043b\u0438\u0447\u0435\u0441\u0442\u0432\u043e \u0443\u0441\u043b\u0443\u0433 \u0438\u043b\u0438 \u0434\u043b\u0438\u043d\u0443 \u043d\u0430\u0437\u0432\u0430\u043d\u0438\u0439.")
}

// EN: Function `renderActPDF`.
//
// EN: What it does: renderActPDF prepares the shared PDF canvas and hands it to the blank assigned to the employee.
//
// EN: Key points: every blank draws on the same A4 page with auto page breaks disabled, and the result is rejected
// EN: when the content spilled past the printable area, so an act is always exactly one sheet.
func renderActPDF(path string, data actPDFData) error {
	template := actTemplateOrDefault(data.ActTemplate)
	fontRegular, fontBold, fontItalic, err := resolvePDFFonts(actTemplateFont(template))
	if err != nil {
		return err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 12, 10)
	pdf.SetAutoPageBreak(false, 0)
	for _, font := range [][2]string{{"", fontRegular}, {"B", fontBold}, {"I", fontItalic}} {
		if font[1] == "" {
			continue
		}
		raw, err := os.ReadFile(font[1])
		if err != nil {
			return fmt.Errorf("read font %s: %w", font[1], err)
		}
		pdf.AddUTF8FontFromBytes("Mercel", font[0], raw)
	}
	if err := pdf.Error(); err != nil {
		return fmt.Errorf("load pdf fonts: %w", err)
	}
	pdf.SetFont("Mercel", "", 10)
	pdf.AddPage()

	var bottom float64
	switch template {
	case ActTemplateFormal:
		bottom, err = renderFormalAct(pdf, data)
	case ActTemplateTypographic:
		bottom, err = renderTypographicAct(pdf, data)
	case ActTemplateContract:
		bottom, err = renderContractAct(pdf, data)
	case ActTemplateTabular:
		bottom, err = renderTabularAct(pdf, data)
	default:
		bottom, err = renderStandardAct(pdf, data)
	}
	if err != nil {
		return err
	}
	if pdf.PageNo() > 1 {
		return errActDoesNotFit()
	}
	if pdf.GetY() > bottom {
		return errActDoesNotFit()
	}
	if err := pdf.OutputFileAndClose(path); err != nil {
		return fmt.Errorf("save pdf: %w", err)
	}
	return nil
}

// EN: Function `renderStandardAct`.
//
// EN: What it does: renderStandardAct draws the standard Mercel blank and reports the bottom edge the content had to stay above.
//
// EN: Key points: the scale is fitted first so the whole act lands on a single page.
func renderStandardAct(pdf *gofpdf.Fpdf, data actPDFData) (float64, error) {
	layout, err := fitActPDFLayout(pdf, data)
	if err != nil {
		return 0, err
	}
	drawActPDF(pdf, data, layout)
	return layout.Bottom, nil
}

// EN: Function `fitActScale`.
//
// EN: What it does: fitActScale finds the largest scale between minimum and target for which the blank still fits.
//
// EN: Key points: shared by every blank; returns false when even the smallest scale overflows the page, which the
// EN: caller turns into a user facing error.
func fitActScale(minimum float64, target float64, fits func(scale float64) bool) (float64, bool) {
	if fits(target) {
		return target, true
	}
	if !fits(minimum) {
		return 0, false
	}
	low := minimum
	high := target
	for i := 0; i < 12; i++ {
		mid := (low + high) / 2
		if fits(mid) {
			low = mid
			continue
		}
		high = mid
	}
	return low, true
}

// EN: Font families used by the act blanks.
//
// EN: What it does: every blank names the typeface it is designed for; the exporter resolves and embeds only that
// EN: one, which both keeps the PDF small and makes the blanks visibly different from each other.
//
// EN: Key points: Windows faces come first, then the Linux/macOS fallbacks used by build and test machines; a
// EN: family that is not installed degrades to its fallback family instead of failing the export.
const (
	actFontSans    = "sans"
	actFontSerif   = "serif"
	actFontGeorgia = "georgia"
	actFontTahoma  = "tahoma"
)

// EN: Variable `actFontCandidates`.
//
// EN: What it does: actFontCandidates lists the regular/bold/italic file trios of every font family.
//
// EN: Key points: an empty italic slot means the family ships without one (Tahoma), and the italic style is then
// EN: simply not registered — no blank that uses such a family asks for italics.
var actFontCandidates = map[string][][3]string{
	actFontSans: {
		{`C:\Windows\Fonts\arial.ttf`, `C:\Windows\Fonts\arialbd.ttf`, `C:\Windows\Fonts\ariali.ttf`},
		{`C:\Windows\Fonts\segoeui.ttf`, `C:\Windows\Fonts\segoeuib.ttf`, `C:\Windows\Fonts\segoeuii.ttf`},
		{"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf", "/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf", "/usr/share/fonts/truetype/liberation/LiberationSans-Italic.ttf"},
		{"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-Oblique.ttf"},
		{"/Library/Fonts/Arial.ttf", "/Library/Fonts/Arial Bold.ttf", "/Library/Fonts/Arial Italic.ttf"},
	},
	actFontSerif: {
		{`C:\Windows\Fonts\times.ttf`, `C:\Windows\Fonts\timesbd.ttf`, `C:\Windows\Fonts\timesi.ttf`},
		{"/usr/share/fonts/truetype/liberation/LiberationSerif-Regular.ttf", "/usr/share/fonts/truetype/liberation/LiberationSerif-Bold.ttf", "/usr/share/fonts/truetype/liberation/LiberationSerif-Italic.ttf"},
		{"/usr/share/fonts/truetype/dejavu/DejaVuSerif.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSerif-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSerif-Italic.ttf"},
		{"/Library/Fonts/Times New Roman.ttf", "/Library/Fonts/Times New Roman Bold.ttf", "/Library/Fonts/Times New Roman Italic.ttf"},
	},
	actFontGeorgia: {
		{`C:\Windows\Fonts\georgia.ttf`, `C:\Windows\Fonts\georgiab.ttf`, `C:\Windows\Fonts\georgiai.ttf`},
		{"/usr/share/fonts/truetype/dejavu/DejaVuSerif.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSerif-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSerif-Italic.ttf"},
		{"/Library/Fonts/Georgia.ttf", "/Library/Fonts/Georgia Bold.ttf", "/Library/Fonts/Georgia Italic.ttf"},
	},
	actFontTahoma: {
		{`C:\Windows\Fonts\tahoma.ttf`, `C:\Windows\Fonts\tahomabd.ttf`, ""},
		{`C:\Windows\Fonts\verdana.ttf`, `C:\Windows\Fonts\verdanab.ttf`, `C:\Windows\Fonts\verdanai.ttf`},
		{"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-Oblique.ttf"},
	},
}

// EN: Variable `actFontFallbacks`.
//
// EN: What it does: actFontFallbacks says which family to try next when the requested one is missing.
//
// EN: Key points: sans has no fallback — failing to find it is the only case that aborts the export.
var actFontFallbacks = map[string]string{
	actFontSerif:   actFontSans,
	actFontGeorgia: actFontSerif,
	actFontTahoma:  actFontSans,
}

// EN: Function `actTemplateFont`.
//
// EN: What it does: actTemplateFont maps a blank onto the font family it is typeset with.
//
// EN: Key points: the typeface is a large part of what tells the blanks apart on paper.
func actTemplateFont(template string) string {
	switch template {
	case ActTemplateFormal:
		return actFontSerif
	case ActTemplateContract:
		return actFontGeorgia
	case ActTemplateTabular:
		return actFontTahoma
	default:
		return actFontSans
	}
}

// EN: Function `resolvePDFFonts`.
//
// EN: What it does: resolvePDFFonts finds an installed regular/bold/italic trio for one font family.
//
// EN: Key points: the returned italic path is empty when the family has no italic face; a missing family falls back
// EN: to a related one, and only a missing sans family is reported as an error.
func resolvePDFFonts(family string) (string, string, string, error) {
	for _, candidate := range actFontCandidates[family] {
		if !fileExists(candidate[0]) || !fileExists(candidate[1]) {
			continue
		}
		italic := candidate[2]
		if italic != "" && !fileExists(italic) {
			italic = ""
		}
		return candidate[0], candidate[1], italic, nil
	}
	if windir := os.Getenv("WINDIR"); windir != "" {
		names := [3]string{"arial.ttf", "arialbd.ttf", "ariali.ttf"}
		if family == actFontSerif {
			names = [3]string{"times.ttf", "timesbd.ttf", "timesi.ttf"}
		}
		regular := filepath.Join(windir, "Fonts", names[0])
		bold := filepath.Join(windir, "Fonts", names[1])
		italic := filepath.Join(windir, "Fonts", names[2])
		if fileExists(regular) && fileExists(bold) {
			if !fileExists(italic) {
				italic = ""
			}
			return regular, bold, italic, nil
		}
	}
	if fallback, ok := actFontFallbacks[family]; ok {
		return resolvePDFFonts(fallback)
	}
	return "", "", "", errors.New("\u041d\u0435 \u0443\u0434\u0430\u043b\u043e\u0441\u044c \u043d\u0430\u0439\u0442\u0438 \u0441\u0438\u0441\u0442\u0435\u043c\u043d\u044b\u0435 \u0448\u0440\u0438\u0444\u0442\u044b \u0434\u043b\u044f \u0433\u0435\u043d\u0435\u0440\u0430\u0446\u0438\u0438 PDF. \u0423\u0431\u0435\u0434\u0438\u0442\u0435\u0441\u044c, \u0447\u0442\u043e \u0432 Windows \u0434\u043e\u0441\u0442\u0443\u043f\u043d\u044b Arial \u0438\u043b\u0438 Segoe UI.")
}

// EN: Function `drawActParagraph`.
//
// EN: What it does: drawActParagraph lays out one paragraph of a blank and returns its bottom edge.
//
// EN: Key points: shared by the blanks that measure and draw in a single walk — the height always comes from
// EN: SplitText, so the measuring pass and the drawing pass can never disagree.
func drawActParagraph(pdf *gofpdf.Fpdf, left float64, width float64, y float64, text string, style string, fontSize float64, lineHeight float64, align string, render bool) float64 {
	pdf.SetFont("Mercel", style, fontSize)
	lines := pdf.SplitText(text, width)
	if len(lines) == 0 {
		lines = []string{""}
	}
	if render {
		pdf.SetXY(left, y)
		pdf.MultiCell(width, lineHeight, text, "", align, false)
	}
	return y + (float64(len(lines)) * lineHeight)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fitActPDFLayout(pdf *gofpdf.Fpdf, data actPDFData) (actPDFLayout, error) {
	// Prefer a larger print-friendly layout, but back off just enough to keep A4 on one page.
	scale, ok := fitActScale(actPDFMinScale, actPDFTargetScale, func(scale float64) bool {
		candidate := newActPDFLayout(pdf, scale)
		return estimateActPDFHeight(pdf, data, candidate) <= candidate.Bottom
	})
	if !ok {
		return actPDFLayout{}, errActDoesNotFit()
	}
	return newActPDFLayout(pdf, scale), nil
}

func newActPDFLayout(pdf *gofpdf.Fpdf, scale float64) actPDFLayout {
	pageWidth, pageHeight := pdf.GetPageSize()
	left := 10.0
	right := 10.0
	contentWidth := pageWidth - left - right
	colNo := 10.0
	colQty := 31.0
	colMoney := 38.0
	colName := contentWidth - colNo - colQty - colMoney

	return actPDFLayout{
		Scale:                 scale,
		Left:                  left,
		Right:                 right,
		Top:                   12.0,
		Bottom:                pageHeight - 10.0,
		ContentWidth:          contentWidth,
		TitleFont:             10.0 * scale,
		MetaFont:              9.0 * scale,
		BodyFont:              8.8 * scale,
		TotalFont:             8.0 * scale,
		ClosingFont:           8.6 * scale,
		SignatureFont:         8.8 * scale,
		TableHeaderFont:       8.8 * scale,
		TableBodyFont:         8.4 * scale,
		TableTotalFont:        9.0 * scale,
		TitleHeight:           6.0 * scale,
		MetaHeight:            6.0 * scale,
		IntroLineHeight:       4.7 * scale,
		TotalLineHeight:       4.2 * scale,
		ClosingLineHeight:     4.8 * scale,
		SignatureHeaderHeight: 5.0 * scale,
		SignatureHeight:       16.0 * scale,
		IntroGap:              5.2 * scale,
		TableGap:              2.0 * scale,
		TotalGap:              4.0 * scale,
		ClosingGap:            5.0 * scale,
		SignatureGap:          6.0 * scale,
		TableHeaderHeight:     8.0 * scale,
		TableTotalHeight:      10.0 * scale,
		TableLineHeight:       4.9 * scale,
		TablePaddingTop:       2.1 * scale,
		TablePaddingX:         2.0 * scale,
		ColNo:                 colNo,
		ColName:               colName,
		ColQty:                colQty,
		ColMoney:              colMoney,
	}
}

func estimateActPDFHeight(pdf *gofpdf.Fpdf, data actPDFData, layout actPDFLayout) float64 {
	y := layout.Top
	y += layout.TitleHeight
	y += layout.MetaHeight
	y += layout.IntroGap
	pdf.SetFont("Mercel", "", layout.BodyFont)
	y += float64(len(pdf.SplitText(buildActIntro(data), layout.ContentWidth))) * layout.IntroLineHeight
	y += layout.TableGap

	rowHeights := actRowHeights(pdf, data.Items, layout)
	y += layout.TableHeaderHeight
	for _, height := range rowHeights {
		y += height
	}
	y += layout.TableTotalHeight

	y += layout.TotalGap
	pdf.SetFont("Mercel", "I", layout.TotalFont)
	y += float64(len(pdf.SplitText(actTotalText(data), layout.ContentWidth))) * layout.TotalLineHeight
	y += layout.ClosingGap
	pdf.SetFont("Mercel", "", layout.ClosingFont)
	y += float64(len(pdf.SplitText(actClosingText(), layout.ContentWidth))) * layout.ClosingLineHeight
	y += layout.SignatureGap
	y += layout.SignatureHeaderHeight
	y += layout.SignatureHeight
	return y
}

func drawActPDF(pdf *gofpdf.Fpdf, data actPDFData, layout actPDFLayout) {
	left := layout.Left
	top := layout.Top
	contentWidth := layout.ContentWidth

	pdf.SetXY(left, top)
	pdf.SetFont("Mercel", "B", layout.TitleFont)
	pdf.CellFormat(contentWidth, layout.TitleHeight, fmt.Sprintf("\u0410\u041a\u0422 \u2116 %d \u041f\u0420\u0418\u0401\u041c\u0410-\u0421\u0414\u0410\u0427\u0418 \u041e\u041a\u0410\u0417\u0410\u041d\u041d\u042b\u0425 \u0423\u0421\u041b\u0423\u0413", data.ActNumber), "", 1, "C", false, 0, "")

	pdf.SetFont("Mercel", "", layout.MetaFont)
	pdf.SetX(left)
	pdf.CellFormat(contentWidth/2, layout.MetaHeight, "\u0433. \u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433", "", 0, "L", false, 0, "")
	pdf.CellFormat(contentWidth/2, layout.MetaHeight, russianActDate(data.GeneratedAt), "", 1, "R", false, 0, "")

	pdf.Ln(layout.IntroGap)
	pdf.SetX(left)
	pdf.SetFont("Mercel", "", layout.BodyFont)
	pdf.MultiCell(contentWidth, layout.IntroLineHeight, buildActIntro(data), "", "J", false)
	pdf.Ln(layout.TableGap)

	drawActTable(pdf, layout, data)

	pdf.Ln(layout.TotalGap)
	pdf.SetX(left)
	pdf.SetFont("Mercel", "I", layout.TotalFont)
	pdf.MultiCell(contentWidth, layout.TotalLineHeight, actTotalText(data), "", "L", false)

	pdf.Ln(layout.ClosingGap)
	pdf.SetX(left)
	pdf.SetFont("Mercel", "", layout.ClosingFont)
	pdf.MultiCell(contentWidth, layout.ClosingLineHeight, actClosingText(), "", "J", false)

	pdf.Ln(layout.SignatureGap)
	signatureY := pdf.GetY()
	colWidth := contentWidth / 2

	pdf.SetFont("Mercel", "B", layout.TableTotalFont)
	pdf.SetXY(left, signatureY)
	pdf.CellFormat(colWidth, layout.SignatureHeaderHeight, "\u0418\u0441\u043f\u043e\u043b\u043d\u0438\u0442\u0435\u043b\u044c:", "", 0, "L", false, 0, "")
	pdf.CellFormat(colWidth, layout.SignatureHeaderHeight, "\u0417\u0430\u043a\u0430\u0437\u0447\u0438\u043a:", "", 1, "L", false, 0, "")

	pdf.SetFont("Mercel", "", layout.SignatureFont)
	pdf.SetX(left)
	drawSignature(pdf, layout, colWidth, shortEmployeeSignatureName(data.EmployeeFullName))
	drawSignature(pdf, layout, colWidth, resolveContractDirectorShort(data.ContractTitle))
	pdf.SetY(signatureY + layout.SignatureHeaderHeight + layout.SignatureHeight)
}

func buildActIntro(data actPDFData) string {
	return fmt.Sprintf(
		"\u041e\u0431\u0449\u0435\u0441\u0442\u0432\u043e \u0441 \u043e\u0433\u0440\u0430\u043d\u0438\u0447\u0435\u043d\u043d\u043e\u0439 \u043e\u0442\u0432\u0435\u0442\u0441\u0442\u0432\u0435\u043d\u043d\u043e\u0441\u0442\u044c\u044e \u00ab%s\u00bb, \u0438\u043c\u0435\u043d\u0443\u0435\u043c\u043e\u0435 \u0432 \u0434\u0430\u043b\u044c\u043d\u0435\u0439\u0448\u0435\u043c \u0417\u0430\u043a\u0430\u0437\u0447\u0438\u043a, \u0432 \u043b\u0438\u0446\u0435 \u0433\u0435\u043d\u0435\u0440\u0430\u043b\u044c\u043d\u043e\u0433\u043e \u0434\u0438\u0440\u0435\u043a\u0442\u043e\u0440\u0430 %s \u0441 \u043e\u0434\u043d\u043e\u0439 \u0441\u0442\u043e\u0440\u043e\u043d\u044b \u0438 \u0418\u043d\u0434\u0438\u0432\u0438\u0434\u0443\u0430\u043b\u044c\u043d\u044b\u0439 \u043f\u0440\u0435\u0434\u043f\u0440\u0438\u043d\u0438\u043c\u0430\u0442\u0435\u043b\u044c %s, \u0438\u043c\u0435\u043d\u0443\u0435\u043c\u044b\u0439 \u0432 \u0434\u0430\u043b\u044c\u043d\u0435\u0439\u0448\u0435\u043c \u0418\u0441\u043f\u043e\u043b\u043d\u0438\u0442\u0435\u043b\u044c, \u0441 \u0434\u0440\u0443\u0433\u043e\u0439 \u0441\u0442\u043e\u0440\u043e\u043d\u044b \u0441\u043e\u0441\u0442\u0430\u0432\u0438\u043b\u0438 \u043d\u0430\u0441\u0442\u043e\u044f\u0449\u0438\u0439 \u0430\u043a\u0442, \u0441\u043e\u0433\u043b\u0430\u0441\u043d\u043e \u0414\u043e\u0433\u043e\u0432\u043e\u0440\u0443 %s \u2116 %s \u043e\u0442 %s, \u0437\u0430\u043a\u043b\u044e\u0447\u0435\u043d\u043d\u043e\u0433\u043e \u043c\u0435\u0436\u0434\u0443 \u0421\u0442\u043e\u0440\u043e\u043d\u0430\u043c\u0438, \u043e \u0442\u043e\u043c, \u0447\u0442\u043e \u0418\u0441\u043f\u043e\u043b\u043d\u0438\u0442\u0435\u043b\u044c \u0432\u044b\u043f\u043e\u043b\u043d\u0438\u043b, \u0430 \u0417\u0430\u043a\u0430\u0437\u0447\u0438\u043a \u043f\u0440\u0438\u043d\u044f\u043b \u0441\u043b\u0435\u0434\u0443\u044e\u0449\u0438\u0435 \u0440\u0430\u0431\u043e\u0442\u044b:",
		data.CustomerName,
		resolveContractDirectorName(data.ContractTitle),
		data.EmployeeFullName,
		data.ContractTitle,
		data.ContractNumber,
		formatContractDateLong(data.ContractDate),
	)
}

func actTotalText(data actPDFData) string {
	return fmt.Sprintf("\u0412\u0441\u0435\u0433\u043e \u0432\u044b\u043f\u043e\u043b\u043d\u0435\u043d\u043e \u0440\u0430\u0431\u043e\u0442 \u043d\u0430 \u0441\u0443\u043c\u043c\u0443: %s %s 00 \u043a\u043e\u043f\u0435\u0435\u043a \u0431\u0435\u0437 \u041d\u0414\u0421.", data.TotalWords, data.TotalCurrency)
}

func actClosingText() string {
	return "\u0412\u044b\u0448\u0435\u043f\u0435\u0440\u0435\u0447\u0438\u0441\u043b\u0435\u043d\u043d\u044b\u0435 \u0443\u0441\u043b\u0443\u0433\u0438 \u0432\u044b\u043f\u043e\u043b\u043d\u0435\u043d\u044b \u043f\u043e\u043b\u043d\u043e\u0441\u0442\u044c\u044e \u0438 \u0432 \u0441\u0440\u043e\u043a. \u0417\u0430\u043a\u0430\u0437\u0447\u0438\u043a \u043f\u0440\u0435\u0442\u0435\u043d\u0437\u0438\u0439 \u043f\u043e \u043e\u0431\u044a\u0451\u043c\u0443, \u043a\u0430\u0447\u0435\u0441\u0442\u0432\u0443 \u0438 \u0441\u0440\u043e\u043a\u0430\u043c \u043e\u043a\u0430\u0437\u0430\u043d\u0438\u044f \u0443\u0441\u043b\u0443\u0433 \u043d\u0435 \u0438\u043c\u0435\u0435\u0442."
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

func drawSignature(pdf *gofpdf.Fpdf, layout actPDFLayout, width float64, name string) {
	nameWidth := width - (28 * layout.Scale)
	if nameWidth < 60 {
		nameWidth = width * 0.7
	}
	pdf.CellFormat(nameWidth, layout.SignatureHeight, name, "", 0, "L", false, 0, "")
	pdf.CellFormat(width-nameWidth, layout.SignatureHeight, "   /________/", "", 0, "L", false, 0, "")
}

func drawActTable(pdf *gofpdf.Fpdf, layout actPDFLayout, data actPDFData) {
	left := layout.Left
	colNo := layout.ColNo
	colName := layout.ColName
	colQty := layout.ColQty
	colMoney := layout.ColMoney
	lineHeight := layout.TableLineHeight
	paddingTop := layout.TablePaddingTop
	paddingHorizontal := layout.TablePaddingX

	pdf.SetFont("Mercel", "B", layout.TableHeaderFont)
	pdf.SetX(left)
	pdf.CellFormat(colNo, layout.TableHeaderHeight, "\u2116", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colName, layout.TableHeaderHeight, "\u041d\u0430\u0438\u043c\u0435\u043d\u043e\u0432\u0430\u043d\u0438\u0435 \u0443\u0441\u043b\u0443\u0433\u0438", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colQty, layout.TableHeaderHeight, "\u041a\u043e\u043b\u0438\u0447\u0435\u0441\u0442\u0432\u043e", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colMoney, layout.TableHeaderHeight, "\u0421\u0442\u043e\u0438\u043c\u043e\u0441\u0442\u044c", "1", 1, "C", false, 0, "")

	pdf.SetFont("Mercel", "", layout.TableBodyFont)
	for index, item := range data.Items {
		lines := pdf.SplitText(item.Name, colName-(paddingHorizontal*2))
		if len(lines) == 0 {
			lines = []string{""}
		}
		rowHeight := float64(len(lines))*lineHeight + (paddingTop * 2)
		if minHeight := 10 * layout.Scale; rowHeight < minHeight {
			rowHeight = minHeight
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

	pdf.SetFont("Mercel", "B", layout.TableTotalFont)
	pdf.SetX(left + colNo + colName)
	pdf.CellFormat(colQty, layout.TableTotalHeight, "\u0418\u0422\u041e\u0413\u041e:", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colMoney, layout.TableTotalHeight, formatDocumentMoney(data.TotalAmount), "1", 1, "C", false, 0, "")
}

func actRowHeights(pdf *gofpdf.Fpdf, items []CalculationItem, layout actPDFLayout) []float64 {
	heights := make([]float64, 0, len(items))
	pdf.SetFont("Mercel", "", layout.TableBodyFont)
	for _, item := range items {
		lines := pdf.SplitText(item.Name, layout.ColName-(layout.TablePaddingX*2))
		if len(lines) == 0 {
			lines = []string{""}
		}
		height := float64(len(lines))*layout.TableLineHeight + (layout.TablePaddingTop * 2)
		if minHeight := 10 * layout.Scale; height < minHeight {
			height = minHeight
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

func rubleNoun(n int) string {
	lastTwo := n % 100
	last := n % 10
	if lastTwo >= 11 && lastTwo <= 14 {
		return "рублей"
	}
	switch last {
	case 1:
		return "рубль"
	case 2, 3, 4:
		return "рубля"
	default:
		return "рублей"
	}
}
