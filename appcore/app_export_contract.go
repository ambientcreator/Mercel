package appcore

import (
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// EN: Contract-style act blank (Акт_бланк_4_договорный).
//
// EN: What it does: this file renders the contract flavoured variant — a requisites header with both parties, a
// EN: Georgia title under a rule, numbered clauses instead of loose paragraphs and a price list framed by double
// EN: rules.
//
// EN: Key points: the source Word blank spills over two pages, so the document is scaled down until it fits a
// EN: single A4 sheet; measuring and drawing share one walk, which keeps the fit exact.
const (
	contractActTargetScale = 1.00
	contractActMinScale    = 0.62
)

type contractActLayout struct {
	Scale        float64
	Left         float64
	Top          float64
	Bottom       float64
	ContentWidth float64

	PartyLabelFont     float64
	PartyLabelTracking float64
	PartyLabelHeight   float64
	PartyLabelGap      float64
	PartyLineFont      float64
	PartyLineHeight    float64
	PartyGutter        float64
	PartyGap           float64

	TitleFont     float64
	TitleTracking float64
	TitleHeight   float64
	TitleRuleGap  float64
	TitleGap      float64

	MetaFont   float64
	MetaHeight float64
	MetaGap    float64

	ClauseFont       float64
	ClauseLineHeight float64
	ClauseGap        float64

	TableHeaderFont   float64
	TableHeaderHeight float64
	TableBodyFont     float64
	TableLineHeight   float64
	TableRowPadY      float64
	TableCellPadX     float64
	TableRowMinHeight float64
	TableTotalFont    float64
	TableTotalValue   float64
	TableTotalHeight  float64
	TableGap          float64

	SignLabelFont     float64
	SignLabelHeight   float64
	SignLabelGap      float64
	SignNameFont      float64
	SignNameHeight    float64
	SignRuleGap       float64
	SignRuleWidth     float64
	SignCaptionFont   float64
	SignCaptionGap    float64
	SignCaptionHeight float64
	SignGap           float64

	ColNo    float64
	ColName  float64
	ColQty   float64
	ColPrice float64
	ColSum   float64

	RuleBold   float64
	RuleThin   float64
	RuleHair   float64
	DoubleGap  float64
	TitleRuleW float64
}

// EN: Function `renderContractAct`.
//
// EN: What it does: renderContractAct fits the contract blank to one page and draws it, returning the bottom edge.
//
// EN: Key points: mirrors the other renderers so the shared page checks in renderActPDF apply here too.
func renderContractAct(pdf *gofpdf.Fpdf, data actPDFData) (float64, error) {
	layout, err := fitContractActLayout(pdf, data)
	if err != nil {
		return 0, err
	}
	layoutContractAct(pdf, data, layout, true)
	return layout.Bottom, nil
}

// EN: Function `fitContractActLayout`.
//
// EN: What it does: fitContractActLayout picks the largest scale that keeps the contract blank on a single sheet.
//
// EN: Key points: this blank carries the most content of all five, so it is allowed to shrink the furthest before
// EN: the export is refused.
func fitContractActLayout(pdf *gofpdf.Fpdf, data actPDFData) (contractActLayout, error) {
	scale, ok := fitActScale(contractActMinScale, contractActTargetScale, func(scale float64) bool {
		candidate := newContractActLayout(pdf, scale)
		return layoutContractAct(pdf, data, candidate, false) <= candidate.Bottom
	})
	if !ok {
		return contractActLayout{}, errActDoesNotFit()
	}
	return newContractActLayout(pdf, scale), nil
}

// EN: Function `newContractActLayout`.
//
// EN: What it does: newContractActLayout converts the Word measurements of the contract blank into millimetres for one scale.
//
// EN: Key points: sizes come from the source document at scale 1.0 (10 pt clauses, 22 mm side margins); margins
// EN: shrink with the content but never below the ones used by the standard blank.
func newContractActLayout(pdf *gofpdf.Fpdf, scale float64) contractActLayout {
	pageWidth, pageHeight := pdf.GetPageSize()

	left := 22.0 * scale
	if left < 12.0 {
		left = 12.0
	}
	top := 17.6 * scale
	if top < 11.0 {
		top = 11.0
	}
	bottomMargin := 20.0 * scale
	if bottomMargin < 10.0 {
		bottomMargin = 10.0
	}
	contentWidth := pageWidth - (left * 2)

	// Column proportions of the source table: 600 / 4238 / 1300 / 1500 / 2000 twips.
	colNo := contentWidth * 0.0623
	colQty := contentWidth * 0.1349
	colPrice := contentWidth * 0.1556
	colSum := contentWidth * 0.2075
	colName := contentWidth - colNo - colQty - colPrice - colSum

	return contractActLayout{
		Scale:        scale,
		Left:         left,
		Top:          top,
		Bottom:       pageHeight - bottomMargin,
		ContentWidth: contentWidth,

		PartyLabelFont:     8.0 * scale,
		PartyLabelTracking: 0.28 * scale,
		PartyLabelHeight:   3.8 * scale,
		PartyLabelGap:      2.8 * scale,
		PartyLineFont:      9.0 * scale,
		PartyLineHeight:    4.6 * scale,
		PartyGutter:        contentWidth * 0.0424,
		PartyGap:           6.0 * scale,

		TitleFont:     12.0 * scale,
		TitleTracking: 0.35 * scale,
		TitleHeight:   6.0 * scale,
		TitleRuleGap:  2.8 * scale,
		TitleGap:      4.6 * scale,

		MetaFont:   9.5 * scale,
		MetaHeight: 4.6 * scale,
		MetaGap:    6.7 * scale,

		ClauseFont:       10.0 * scale,
		ClauseLineHeight: 5.3 * scale,
		ClauseGap:        4.2 * scale,

		TableHeaderFont:   8.5 * scale,
		TableHeaderHeight: 7.4 * scale,
		TableBodyFont:     9.0 * scale,
		TableLineHeight:   4.4 * scale,
		TableRowPadY:      1.4 * scale,
		TableCellPadX:     1.9 * scale,
		TableRowMinHeight: 7.8 * scale,
		TableTotalFont:    9.0 * scale,
		TableTotalValue:   10.0 * scale,
		TableTotalHeight:  8.8 * scale,
		TableGap:          4.2 * scale,

		SignLabelFont:     8.5 * scale,
		SignLabelHeight:   4.4 * scale,
		SignLabelGap:      6.0 * scale,
		SignNameFont:      9.0 * scale,
		SignNameHeight:    4.6 * scale,
		SignRuleGap:       1.6 * scale,
		SignRuleWidth:     46.0 * scale,
		SignCaptionFont:   7.0 * scale,
		SignCaptionGap:    1.4 * scale,
		SignCaptionHeight: 3.2 * scale,
		SignGap:           5.3 * scale,

		ColNo:    colNo,
		ColName:  colName,
		ColQty:   colQty,
		ColPrice: colPrice,
		ColSum:   colSum,

		RuleBold:   0.35,
		RuleThin:   0.20,
		RuleHair:   0.10,
		DoubleGap:  0.7,
		TitleRuleW: 0.44,
	}
}

// EN: Function `layoutContractAct`.
//
// EN: What it does: layoutContractAct walks the contract blank once and either measures it or draws it.
//
// EN: Key points: measuring and drawing share this walk, so the fitted scale always matches what ends up on the
// EN: page; the function returns the y coordinate right below the last element.
func layoutContractAct(pdf *gofpdf.Fpdf, data actPDFData, layout contractActLayout, render bool) float64 {
	left := layout.Left
	width := layout.ContentWidth
	y := layout.Top

	y = drawContractParties(pdf, layout, data, y, render) + layout.PartyGap

	if render {
		pdf.SetFont("Mercel", "B", layout.TitleFont)
		title := fmt.Sprintf("АКТ № %d ПРИЁМА-СДАЧИ ОКАЗАННЫХ УСЛУГ", data.ActNumber)
		titleWidth := trackedTextWidth(pdf, title, layout.TitleTracking)
		drawTrackedText(pdf, left+((width-titleWidth)/2), y, layout.TitleHeight, title, layout.TitleTracking)
	}
	y += layout.TitleHeight + layout.TitleRuleGap
	if render {
		pdf.SetLineWidth(layout.TitleRuleW)
		pdf.Line(left, y, left+width, y)
	}
	y += layout.TitleGap

	if render {
		pdf.SetFont("Mercel", "", layout.MetaFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(width/2, layout.MetaHeight, "г. Санкт-Петербург", "", 0, "L", false, 0, "")
		pdf.CellFormat(width/2, layout.MetaHeight, russianActDate(data.GeneratedAt), "", 0, "R", false, 0, "")
	}
	y += layout.MetaHeight + layout.MetaGap

	y = drawActParagraph(pdf, left, width, y, contractActClause1(data), "", layout.ClauseFont, layout.ClauseLineHeight, "J", render) + layout.ClauseGap
	y = drawContractActTable(pdf, layout, data, y, render) + layout.TableGap

	for _, clause := range []string{contractActClause2(data), contractActClause3(), contractActClause4()} {
		y = drawActParagraph(pdf, left, width, y, clause, "", layout.ClauseFont, layout.ClauseLineHeight, "J", render) + layout.ClauseGap
	}

	y += layout.SignGap
	y = drawContractSignatures(pdf, layout, data, y, render)
	return y
}

// EN: Function `drawContractParties`.
//
// EN: What it does: drawContractParties renders the header naming both parties side by side.
//
// EN: Key points: the source blank also carries ИНН/ОГРНИП/КПП lines, but the contract is already signed by the
// EN: time the act is issued, so only the names are printed.
func drawContractParties(pdf *gofpdf.Fpdf, layout contractActLayout, data actPDFData, y float64, render bool) float64 {
	columnWidth := (layout.ContentWidth - layout.PartyGutter) / 2
	leftColumn := layout.Left
	rightColumn := layout.Left + columnWidth + layout.PartyGutter

	contractor := []string{fmt.Sprintf("ИП %s", strings.TrimSpace(data.EmployeeFullName))}
	customer := []string{fmt.Sprintf("ООО «%s»", data.CustomerName)}

	bottom := drawContractPartyColumn(pdf, layout, leftColumn, columnWidth, y, "ИСПОЛНИТЕЛЬ", contractor, render)
	if right := drawContractPartyColumn(pdf, layout, rightColumn, columnWidth, y, "ЗАКАЗЧИК", customer, render); right > bottom {
		bottom = right
	}
	return bottom
}

// EN: Function `drawContractPartyColumn`.
//
// EN: What it does: drawContractPartyColumn stacks the label and the requisite lines of one party.
//
// EN: Key points: long company names wrap inside the column, so the height is measured, not assumed.
func drawContractPartyColumn(pdf *gofpdf.Fpdf, layout contractActLayout, x float64, width float64, y float64, label string, lines []string, render bool) float64 {
	if render {
		pdf.SetFont("Mercel", "B", layout.PartyLabelFont)
		drawTrackedText(pdf, x, y, layout.PartyLabelHeight, label, layout.PartyLabelTracking)
	}
	cursor := y + layout.PartyLabelHeight + layout.PartyLabelGap

	pdf.SetFont("Mercel", "", layout.PartyLineFont)
	for _, line := range lines {
		cursor = drawActParagraph(pdf, x, width, cursor, line, "", layout.PartyLineFont, layout.PartyLineHeight, "L", render)
	}
	return cursor
}

// EN: Function `drawContractActTable`.
//
// EN: What it does: drawContractActTable renders the five column price list of the contract blank.
//
// EN: Key points: the table is framed by double rules on top and around the total, with hairlines between the rows
// EN: and no vertical lines at all, exactly like the source blank.
func drawContractActTable(pdf *gofpdf.Fpdf, layout contractActLayout, data actPDFData, y float64, render bool) float64 {
	left := layout.Left
	right := left + layout.ContentWidth
	padX := layout.TableCellPadX

	xNo := left
	xName := xNo + layout.ColNo
	xQty := xName + layout.ColName
	xPrice := xQty + layout.ColQty
	xSum := xPrice + layout.ColPrice

	if render {
		drawDoubleRule(pdf, left, right, y, layout.RuleBold, layout.DoubleGap)
		pdf.SetFont("Mercel", "B", layout.TableHeaderFont)
		pdf.SetXY(xNo, y)
		pdf.CellFormat(layout.ColNo, layout.TableHeaderHeight, "№", "", 0, "C", false, 0, "")
		pdf.SetXY(xName+padX, y)
		pdf.CellFormat(layout.ColName-padX, layout.TableHeaderHeight, "Наименование услуги", "", 0, "L", false, 0, "")
		pdf.SetXY(xQty, y)
		pdf.CellFormat(layout.ColQty, layout.TableHeaderHeight, "Кол-во", "", 0, "C", false, 0, "")
		pdf.SetXY(xPrice, y)
		pdf.CellFormat(layout.ColPrice-padX, layout.TableHeaderHeight, "Цена", "", 0, "R", false, 0, "")
		pdf.SetXY(xSum, y)
		pdf.CellFormat(layout.ColSum-padX, layout.TableHeaderHeight, "Сумма, руб.", "", 0, "R", false, 0, "")
	}
	y += layout.TableHeaderHeight
	if render {
		pdf.SetLineWidth(layout.RuleThin)
		pdf.Line(left, y, right, y)
	}

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
			pdf.SetXY(xNo, y)
			pdf.CellFormat(layout.ColNo, rowHeight, fmt.Sprintf("%d.", index+1), "", 0, "C", false, 0, "")

			textY := y + ((rowHeight - (float64(len(lines)) * layout.TableLineHeight)) / 2)
			for lineIndex, line := range lines {
				pdf.SetXY(xName+padX, textY+(float64(lineIndex)*layout.TableLineHeight))
				pdf.CellFormat(layout.ColName-(padX*2), layout.TableLineHeight, line, "", 0, "L", false, 0, "")
			}

			pdf.SetXY(xQty, y)
			pdf.CellFormat(layout.ColQty, rowHeight, formatDocumentQuantity(item.Quantity, item.Unit), "", 0, "C", false, 0, "")
			pdf.SetXY(xPrice, y)
			pdf.CellFormat(layout.ColPrice-padX, rowHeight, formatThousands(item.Rate), "", 0, "R", false, 0, "")
			pdf.SetXY(xSum, y)
			pdf.CellFormat(layout.ColSum-padX, rowHeight, formatThousands(item.LineTotal), "", 0, "R", false, 0, "")
		}
		y += rowHeight
		if render {
			pdf.SetLineWidth(layout.RuleHair)
			pdf.Line(left, y, right, y)
		}
	}

	if render {
		drawDoubleRule(pdf, left, right, y, layout.RuleBold, layout.DoubleGap)
		pdf.SetFont("Mercel", "B", layout.TableTotalFont)
		pdf.SetXY(xPrice, y)
		pdf.CellFormat(layout.ColPrice-padX, layout.TableTotalHeight, "Итого:", "", 0, "R", false, 0, "")
		pdf.SetFont("Mercel", "B", layout.TableTotalValue)
		pdf.SetXY(xSum, y)
		pdf.CellFormat(layout.ColSum-padX, layout.TableTotalHeight, formatThousands(data.TotalAmount), "", 0, "R", false, 0, "")
	}
	y += layout.TableTotalHeight
	if render {
		drawDoubleRule(pdf, left, right, y, layout.RuleBold, layout.DoubleGap)
	}
	return y
}

// EN: Function `drawDoubleRule`.
//
// EN: What it does: drawDoubleRule draws the two thin parallel lines used as a double border.
//
// EN: Key points: gofpdf has no double border style, so the pair is drawn by hand around the given baseline.
func drawDoubleRule(pdf *gofpdf.Fpdf, left float64, right float64, y float64, width float64, gap float64) {
	pdf.SetLineWidth(width)
	pdf.Line(left, y-gap, right, y-gap)
	pdf.Line(left, y, right, y)
}

// EN: Function `drawContractSignatures`.
//
// EN: What it does: drawContractSignatures renders the two signature columns closing the contract blank.
//
// EN: Key points: both columns are the same height, so the returned bottom edge is what the fit check compares
// EN: against the printable area.
func drawContractSignatures(pdf *gofpdf.Fpdf, layout contractActLayout, data actPDFData, y float64, render bool) float64 {
	columnWidth := (layout.ContentWidth - layout.PartyGutter) / 2
	leftColumn := layout.Left
	rightColumn := layout.Left + columnWidth + layout.PartyGutter

	bottom := drawContractSignatureColumn(pdf, layout, leftColumn, y, "От Исполнителя", shortEmployeeSignatureName(data.EmployeeFullName), render)
	drawContractSignatureColumn(pdf, layout, rightColumn, y, "От Заказчика", resolveContractDirectorShort(data.ContractTitle), render)
	return bottom
}

// EN: Function `drawContractSignatureColumn`.
//
// EN: What it does: drawContractSignatureColumn stacks the party label, the printed name, the signature rule and its caption.
//
// EN: Key points: the printed name goes above the rule so the exported act is filled in automatically.
func drawContractSignatureColumn(pdf *gofpdf.Fpdf, layout contractActLayout, x float64, y float64, label string, name string, render bool) float64 {
	if render {
		pdf.SetFont("Mercel", "B", layout.SignLabelFont)
		pdf.SetXY(x, y)
		pdf.CellFormat(layout.SignRuleWidth, layout.SignLabelHeight, label, "", 0, "L", false, 0, "")
	}
	cursor := y + layout.SignLabelHeight + layout.SignLabelGap

	if render {
		pdf.SetFont("Mercel", "", layout.SignNameFont)
		pdf.SetXY(x, cursor)
		pdf.CellFormat(layout.SignRuleWidth, layout.SignNameHeight, name, "", 0, "L", false, 0, "")
	}
	cursor += layout.SignNameHeight + layout.SignRuleGap

	if render {
		pdf.SetLineWidth(layout.RuleHair)
		pdf.Line(x, cursor, x+layout.SignRuleWidth, cursor)
	}
	cursor += layout.SignCaptionGap

	if render {
		pdf.SetFont("Mercel", "", layout.SignCaptionFont)
		pdf.SetXY(x, cursor)
		pdf.CellFormat(layout.SignRuleWidth, layout.SignCaptionHeight, "подпись / расшифровка", "", 0, "L", false, 0, "")
	}
	return cursor + layout.SignCaptionHeight
}

// EN: Function `contractActClause1`.
//
// EN: What it does: contractActClause1 builds the opening numbered clause of the contract blank.
//
// EN: Key points: this blank states the scope through numbered clauses instead of one long preamble.
func contractActClause1(data actPDFData) string {
	return fmt.Sprintf(
		"1.  Исполнитель оказал, а Заказчик принял услуги по Договору № %s от %s в следующем объёме:",
		data.ContractNumber,
		formalContractDate(data.ContractDate),
	)
}

// EN: Function `contractActClause2`.
//
// EN: What it does: contractActClause2 spells out the total of the contract blank.
func contractActClause2(data actPDFData) string {
	return fmt.Sprintf("2.  Общая стоимость оказанных услуг составила %s %s 00 копеек, без НДС.", data.TotalWords, data.TotalCurrency)
}

// EN: Function `contractActClause3`.
//
// EN: What it does: contractActClause3 returns the no-claims clause of the contract blank.
func contractActClause3() string {
	return "3.  Услуги оказаны полностью и в срок. Заказчик претензий по объёму, качеству и срокам оказания услуг не имеет."
}

// EN: Function `contractActClause4`.
//
// EN: What it does: contractActClause4 returns the two-copies clause that closes the contract blank.
func contractActClause4() string {
	return "4.  Настоящий акт составлен в двух экземплярах, имеющих равную юридическую силу, по одному для каждой из Сторон."
}
