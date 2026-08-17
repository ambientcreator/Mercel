package appcore

import (
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// EN: Formal act blank (Акт_бланк_1).
//
// EN: What it does: this file renders the typewritten variant of the act — Times New Roman, 20 mm margins, a fully
// EN: framed four column table and signature lines with the name in slashes.
//
// EN: Key points: it deliberately reads as a plain office document, which is what tells it apart from the standard
// EN: blank (sans-serif, tighter margins) and from the typographic one; like every blank it is scaled down until it
// EN: fits a single A4 sheet, and one walk over the document both measures and draws it.
const (
	formalActTargetScale = 1.06
	formalActMinScale    = 0.70
)

type formalActLayout struct {
	Scale        float64
	Left         float64
	Top          float64
	Bottom       float64
	ContentWidth float64

	TitleFont   float64
	TitleHeight float64
	TitleGap    float64

	MetaFont   float64
	MetaHeight float64
	MetaGap    float64

	IntroFont       float64
	IntroLineHeight float64
	IntroGap        float64

	TableHeaderFont   float64
	TableHeaderHeight float64
	TableBodyFont     float64
	TableLineHeight   float64
	TableRowPadY      float64
	TableCellPadX     float64
	TableRowMinHeight float64
	TableTotalFont    float64
	TableTotalHeight  float64
	TableGap          float64

	TotalFont       float64
	TotalLineHeight float64
	TotalGap        float64

	ClosingFont       float64
	ClosingLineHeight float64
	ClosingGap        float64

	SignatureFont       float64
	SignatureHeight     float64
	SignatureGap        float64
	SignatureLineHeight float64
	SignatureColumn     float64

	ColNo    float64
	ColName  float64
	ColQty   float64
	ColMoney float64

	Border float64
}

// EN: Function `renderFormalAct`.
//
// EN: What it does: renderFormalAct fits the formal blank to one page and draws it, returning the bottom edge.
//
// EN: Key points: mirrors the other renderers so the shared page checks in renderActPDF apply here too.
func renderFormalAct(pdf *gofpdf.Fpdf, data actPDFData) (float64, error) {
	layout, err := fitFormalActLayout(pdf, data)
	if err != nil {
		return 0, err
	}
	layoutFormalAct(pdf, data, layout, true)
	return layout.Bottom, nil
}

// EN: Function `fitFormalActLayout`.
//
// EN: What it does: fitFormalActLayout picks the largest scale that keeps the formal blank on a single sheet.
//
// EN: Key points: refuses the export with the shared hint when even the smallest scale overflows.
func fitFormalActLayout(pdf *gofpdf.Fpdf, data actPDFData) (formalActLayout, error) {
	scale, ok := fitActScale(formalActMinScale, formalActTargetScale, func(scale float64) bool {
		candidate := newFormalActLayout(pdf, scale)
		return layoutFormalAct(pdf, data, candidate, false) <= candidate.Bottom
	})
	if !ok {
		return formalActLayout{}, errActDoesNotFit()
	}
	return newFormalActLayout(pdf, scale), nil
}

// EN: Function `newFormalActLayout`.
//
// EN: What it does: newFormalActLayout converts the Word measurements of the formal blank into millimetres for one scale.
//
// EN: Key points: sizes come from the source document at scale 1.0 (11 pt body, 20 mm margins, 470 twip rows);
// EN: margins shrink with the content but never below the ones used by the standard blank.
func newFormalActLayout(pdf *gofpdf.Fpdf, scale float64) formalActLayout {
	pageWidth, pageHeight := pdf.GetPageSize()

	left := 20.0 * scale
	if left < 12.0 {
		left = 12.0
	}
	top := 20.0 * scale
	if top < 12.0 {
		top = 12.0
	}
	bottomMargin := 20.0 * scale
	if bottomMargin < 10.0 {
		bottomMargin = 10.0
	}
	contentWidth := pageWidth - (left * 2)

	// Column proportions of the source table: 620 / 5218 / 1700 / 2100 twips.
	colNo := contentWidth * 0.0643
	colQty := contentWidth * 0.1764
	colMoney := contentWidth * 0.2179
	colName := contentWidth - colNo - colQty - colMoney

	return formalActLayout{
		Scale:        scale,
		Left:         left,
		Top:          top,
		Bottom:       pageHeight - bottomMargin,
		ContentWidth: contentWidth,

		TitleFont:   12.0 * scale,
		TitleHeight: 6.0 * scale,
		TitleGap:    5.3 * scale,

		MetaFont:   11.0 * scale,
		MetaHeight: 5.0 * scale,
		MetaGap:    6.0 * scale,

		IntroFont:       11.0 * scale,
		IntroLineHeight: 5.3 * scale,
		IntroGap:        6.0 * scale,

		TableHeaderFont:   10.0 * scale,
		TableHeaderHeight: 8.3 * scale,
		TableBodyFont:     10.0 * scale,
		TableLineHeight:   4.6 * scale,
		TableRowPadY:      1.4 * scale,
		TableCellPadX:     1.9 * scale,
		TableRowMinHeight: 8.3 * scale,
		TableTotalFont:    10.0 * scale,
		TableTotalHeight:  8.3 * scale,
		TableGap:          5.3 * scale,

		TotalFont:       11.0 * scale,
		TotalLineHeight: 5.3 * scale,
		TotalGap:        6.0 * scale,

		ClosingFont:       11.0 * scale,
		ClosingLineHeight: 5.3 * scale,
		ClosingGap:        11.3 * scale,

		SignatureFont:       11.0 * scale,
		SignatureHeight:     5.0 * scale,
		SignatureGap:        7.4 * scale,
		SignatureLineHeight: 5.0 * scale,
		SignatureColumn:     contentWidth * 0.5292,

		ColNo:    colNo,
		ColName:  colName,
		ColQty:   colQty,
		ColMoney: colMoney,

		Border: 0.26,
	}
}

// EN: Function `layoutFormalAct`.
//
// EN: What it does: layoutFormalAct walks the formal blank once and either measures it or draws it.
//
// EN: Key points: measuring and drawing share this walk, so the fitted scale always matches what ends up on the
// EN: page; the function returns the y coordinate right below the last element.
func layoutFormalAct(pdf *gofpdf.Fpdf, data actPDFData, layout formalActLayout, render bool) float64 {
	left := layout.Left
	width := layout.ContentWidth
	y := layout.Top

	if render {
		pdf.SetFont("Mercel", "B", layout.TitleFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(width, layout.TitleHeight, fmt.Sprintf("АКТ № %d ПРИЁМА-СДАЧИ ОКАЗАННЫХ УСЛУГ", data.ActNumber), "", 0, "C", false, 0, "")
	}
	y += layout.TitleHeight + layout.TitleGap

	if render {
		pdf.SetFont("Mercel", "", layout.MetaFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(width/2, layout.MetaHeight, "г. Санкт-Петербург", "", 0, "L", false, 0, "")
		pdf.CellFormat(width/2, layout.MetaHeight, russianActDate(data.GeneratedAt), "", 0, "R", false, 0, "")
	}
	y += layout.MetaHeight + layout.MetaGap

	y = drawActParagraph(pdf, layout.Left, layout.ContentWidth, y, formalActIntro(data), "", layout.IntroFont, layout.IntroLineHeight, "J", render) + layout.IntroGap

	y = drawFormalActTable(pdf, layout, data, y, render) + layout.TableGap

	y = drawActParagraph(pdf, layout.Left, layout.ContentWidth, y, formalActTotalText(data), "I", layout.TotalFont, layout.TotalLineHeight, "L", render) + layout.TotalGap

	y = drawActParagraph(pdf, layout.Left, layout.ContentWidth, y, actClosingText(), "", layout.ClosingFont, layout.ClosingLineHeight, "J", render) + layout.ClosingGap

	y = drawFormalSignatures(pdf, layout, data, y, render)
	return y
}

// EN: Function `drawFormalActTable`.
//
// EN: What it does: drawFormalActTable renders the fully framed four column table of the formal blank.
//
// EN: Key points: every cell is boxed, and the total row leaves the first two columns empty with only a rule on
// EN: top, exactly like the source blank.
func drawFormalActTable(pdf *gofpdf.Fpdf, layout formalActLayout, data actPDFData, y float64, render bool) float64 {
	left := layout.Left
	padX := layout.TableCellPadX

	xNo := left
	xName := xNo + layout.ColNo
	xQty := xName + layout.ColName
	xMoney := xQty + layout.ColQty

	if render {
		pdf.SetLineWidth(layout.Border)
		pdf.SetFont("Mercel", "B", layout.TableHeaderFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(layout.ColNo, layout.TableHeaderHeight, "№", "1", 0, "C", false, 0, "")
		pdf.CellFormat(layout.ColName, layout.TableHeaderHeight, "Наименование услуги", "1", 0, "C", false, 0, "")
		pdf.CellFormat(layout.ColQty, layout.TableHeaderHeight, "Количество", "1", 0, "C", false, 0, "")
		pdf.CellFormat(layout.ColMoney, layout.TableHeaderHeight, "Стоимость, руб.", "1", 0, "C", false, 0, "")
	}
	y += layout.TableHeaderHeight

	pdf.SetFont("Mercel", "", layout.TableBodyFont)
	for index, item := range data.Items {
		lines := pdf.SplitText(item.Name, layout.ColName-(padX*2))
		if len(lines) == 0 {
			lines = []string{""}
		}
		rowHeight := (float64(len(lines)) * layout.TableLineHeight) + (layout.TableRowPadY * 2)
		if rowHeight < layout.TableRowMinHeight {
			rowHeight = layout.TableRowMinHeight
		}
		if render {
			pdf.SetFont("Mercel", "", layout.TableBodyFont)
			pdf.Rect(xNo, y, layout.ColNo, rowHeight, "")
			pdf.Rect(xName, y, layout.ColName, rowHeight, "")
			pdf.Rect(xQty, y, layout.ColQty, rowHeight, "")
			pdf.Rect(xMoney, y, layout.ColMoney, rowHeight, "")

			pdf.SetXY(xNo, y)
			pdf.CellFormat(layout.ColNo, rowHeight, fmt.Sprintf("%d", index+1), "", 0, "C", false, 0, "")

			textY := y + ((rowHeight - (float64(len(lines)) * layout.TableLineHeight)) / 2)
			for lineIndex, line := range lines {
				pdf.SetXY(xName+padX, textY+(float64(lineIndex)*layout.TableLineHeight))
				pdf.CellFormat(layout.ColName-(padX*2), layout.TableLineHeight, line, "", 0, "L", false, 0, "")
			}

			pdf.SetXY(xQty, y)
			pdf.CellFormat(layout.ColQty, rowHeight, formatDocumentQuantity(item.Quantity, item.Unit), "", 0, "C", false, 0, "")
			pdf.SetXY(xMoney, y)
			pdf.CellFormat(layout.ColMoney, rowHeight, formatThousands(item.LineTotal), "", 0, "C", false, 0, "")
		}
		y += rowHeight
	}

	if render {
		pdf.SetFont("Mercel", "B", layout.TableTotalFont)
		pdf.Line(xNo, y, xQty, y)
		pdf.SetXY(xQty, y)
		pdf.CellFormat(layout.ColQty, layout.TableTotalHeight, "ИТОГО:", "1", 0, "R", false, 0, "")
		pdf.CellFormat(layout.ColMoney, layout.TableTotalHeight, formatThousands(data.TotalAmount), "1", 0, "C", false, 0, "")
	}
	return y + layout.TableTotalHeight
}

// EN: Function `drawFormalSignatures`.
//
// EN: What it does: drawFormalSignatures renders the two signature columns closing the formal blank.
//
// EN: Key points: the second column starts at the tab stop of the source document, and the short name goes into the
// EN: slashes next to the signature line, which is where a signed copy carries it.
func drawFormalSignatures(pdf *gofpdf.Fpdf, layout formalActLayout, data actPDFData, y float64, render bool) float64 {
	leftColumn := layout.Left
	rightColumn := layout.Left + layout.SignatureColumn
	columnWidth := layout.ContentWidth - layout.SignatureColumn

	if render {
		pdf.SetFont("Mercel", "B", layout.SignatureFont)
		pdf.SetXY(leftColumn, y)
		pdf.CellFormat(layout.SignatureColumn, layout.SignatureHeight, "Исполнитель:", "", 0, "L", false, 0, "")
		pdf.SetXY(rightColumn, y)
		pdf.CellFormat(columnWidth, layout.SignatureHeight, "Заказчик:", "", 0, "L", false, 0, "")
	}
	y += layout.SignatureHeight + layout.SignatureGap

	if render {
		pdf.SetFont("Mercel", "", layout.SignatureFont)
		pdf.SetXY(leftColumn, y)
		pdf.CellFormat(layout.SignatureColumn, layout.SignatureLineHeight, formalSignatureLine(shortEmployeeSignatureName(data.EmployeeFullName)), "", 0, "L", false, 0, "")
		pdf.SetXY(rightColumn, y)
		pdf.CellFormat(columnWidth, layout.SignatureLineHeight, formalSignatureLine(resolveContractDirectorShort(data.ContractTitle)), "", 0, "L", false, 0, "")
	}
	return y + layout.SignatureLineHeight
}

// EN: Function `formalSignatureLine`.
//
// EN: What it does: formalSignatureLine builds the "signature line / name" pair used by the formal blank.
//
// EN: Key points: keeps the underscores of the source blank for the handwritten signature and fills in the printed
// EN: name after them.
func formalSignatureLine(name string) string {
	return fmt.Sprintf("________________________  /%s/", strings.TrimSpace(name))
}

// EN: Function `formalActIntro`.
//
// EN: What it does: formalActIntro builds the opening paragraph of the formal blank from the calculation data.
//
// EN: Key points: keeps the wording of the source blank, which unlike the standard one names the basis the customer
// EN: director acts on and refers to the contract by number only.
func formalActIntro(data actPDFData) string {
	return fmt.Sprintf(
		"Общество с ограниченной ответственностью «%s», именуемое в дальнейшем Заказчик, в лице генерального директора %s, действующего на основании Устава, с одной стороны, и Индивидуальный предприниматель %s, именуемый в дальнейшем Исполнитель, с другой стороны, составили настоящий акт согласно Договору № %s от %s, заключённому между Сторонами, о том, что Исполнитель выполнил, а Заказчик принял следующие работы:",
		data.CustomerName,
		resolveContractDirectorName(data.ContractTitle),
		strings.TrimSpace(data.EmployeeFullName),
		data.ContractNumber,
		formalContractDate(data.ContractDate),
	)
}

// EN: Function `formalActTotalText`.
//
// EN: What it does: formalActTotalText spells out the total the way the formal blank does.
//
// EN: Key points: differs from the standard blank by the comma before «без НДС», as in the source document.
func formalActTotalText(data actPDFData) string {
	return fmt.Sprintf("Всего выполнено работ на сумму: %s %s 00 копеек, без НДС.", data.TotalWords, data.TotalCurrency)
}

// EN: Function `formalContractDate`.
//
// EN: What it does: formalContractDate formats the contract date in the «дд» месяц гггг г. form of the formal blank.
//
// EN: Key points: the standard blank writes «года» in full, this one abbreviates it to «г.».
func formalContractDate(value time.Time) string {
	return fmt.Sprintf("«%02d» %s %d г.", value.Day(), russianMonthGenitive(int(value.Month())), value.Year())
}
