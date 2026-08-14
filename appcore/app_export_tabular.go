package appcore

import (
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// EN: Tabular act blank (Акт_бланк_5).
//
// EN: What it does: this file renders the form-like variant — the whole act lives inside one framed grid: title
// EN: cell, labelled rows for the parties, the price list, the total, the amount in words and the signatures.
//
// EN: Key points: Tahoma and the continuous frame are what tell it apart from the other blanks; like every blank it
// EN: is scaled down until it fits a single A4 sheet, and one walk both measures and draws it.
const (
	tabularActTargetScale = 1.00
	tabularActMinScale    = 0.62
)

type tabularActLayout struct {
	Scale        float64
	Left         float64
	Top          float64
	Bottom       float64
	ContentWidth float64

	TitleFont       float64
	TitleTracking   float64
	TitleHeight     float64
	SubtitleFont    float64
	SubtitleHeight  float64
	TitlePadY       float64
	LabelFont       float64
	LabelTracking   float64
	LabelHeight     float64
	ValueFont       float64
	ValueLineHeight float64
	InfoPadY        float64

	HeaderFont     float64
	HeaderTracking float64
	HeaderHeight   float64
	BodyFont       float64
	BodyLineHeight float64
	RowPadY        float64
	RowMinHeight   float64
	CellPadX       float64

	TotalFont      float64
	TotalValueFont float64
	TotalHeight    float64

	WordsPadY      float64
	ClosingFont    float64
	ClosingHeight  float64
	ClosingPadY    float64
	SignLabelGap   float64
	SignNameFont   float64
	SignNameHeight float64
	SignRuleGap    float64
	SignRuleWidth  float64
	SignCaption    float64
	SignCaptionGap float64
	SignCaptionH   float64
	SignPadY       float64

	ColNo    float64
	ColName  float64
	ColQty   float64
	ColPrice float64
	ColSum   float64
	ColLabel float64

	Border float64
}

// EN: Function `renderTabularAct`.
//
// EN: What it does: renderTabularAct fits blank №5 to one page and draws it, returning the bottom edge.
//
// EN: Key points: mirrors the other renderers so the shared page checks in renderActPDF apply here too.
func renderTabularAct(pdf *gofpdf.Fpdf, data actPDFData) (float64, error) {
	layout, err := fitTabularActLayout(pdf, data)
	if err != nil {
		return 0, err
	}
	layoutTabularAct(pdf, data, layout, true)
	return layout.Bottom, nil
}

// EN: Function `fitTabularActLayout`.
//
// EN: What it does: fitTabularActLayout picks the largest scale that keeps blank №5 on a single sheet.
//
// EN: Key points: refuses the export with the shared hint when even the smallest scale overflows.
func fitTabularActLayout(pdf *gofpdf.Fpdf, data actPDFData) (tabularActLayout, error) {
	scale, ok := fitActScale(tabularActMinScale, tabularActTargetScale, func(scale float64) bool {
		candidate := newTabularActLayout(pdf, scale)
		return layoutTabularAct(pdf, data, candidate, false) <= candidate.Bottom
	})
	if !ok {
		return tabularActLayout{}, errActDoesNotFit()
	}
	return newTabularActLayout(pdf, scale), nil
}

// EN: Function `newTabularActLayout`.
//
// EN: What it does: newTabularActLayout converts the Word measurements of blank №5 into millimetres for one scale.
//
// EN: Key points: sizes come from the source document at scale 1.0 (8.5 pt body, 20 mm side margins, 420 twip
// EN: rows); margins shrink with the content but never below the ones used by the standard blank.
func newTabularActLayout(pdf *gofpdf.Fpdf, scale float64) tabularActLayout {
	pageWidth, pageHeight := pdf.GetPageSize()

	left := 20.0 * scale
	if left < 12.0 {
		left = 12.0
	}
	top := 17.6 * scale
	if top < 11.0 {
		top = 11.0
	}
	bottomMargin := 17.6 * scale
	if bottomMargin < 10.0 {
		bottomMargin = 10.0
	}
	contentWidth := pageWidth - (left * 2)

	// Column proportions of the source grid: 520 / 3420 / 1240 / 1660 / 2798 twips.
	colNo := contentWidth * 0.0540
	colQty := contentWidth * 0.1286
	colPrice := contentWidth * 0.1722
	colSum := contentWidth * 0.2903
	colName := contentWidth - colNo - colQty - colPrice - colSum

	return tabularActLayout{
		Scale:        scale,
		Left:         left,
		Top:          top,
		Bottom:       pageHeight - bottomMargin,
		ContentWidth: contentWidth,

		TitleFont:       12.0 * scale,
		TitleTracking:   0.46 * scale,
		TitleHeight:     6.2 * scale,
		SubtitleFont:    8.5 * scale,
		SubtitleHeight:  4.4 * scale,
		TitlePadY:       2.4 * scale,
		LabelFont:       7.0 * scale,
		LabelTracking:   0.25 * scale,
		LabelHeight:     3.4 * scale,
		ValueFont:       9.0 * scale,
		ValueLineHeight: 4.6 * scale,
		InfoPadY:        1.6 * scale,

		HeaderFont:     7.5 * scale,
		HeaderTracking: 0.2 * scale,
		HeaderHeight:   6.4 * scale,
		BodyFont:       8.5 * scale,
		BodyLineHeight: 4.2 * scale,
		RowPadY:        1.4 * scale,
		RowMinHeight:   7.4 * scale,
		CellPadX:       1.9 * scale,

		TotalFont:      8.5 * scale,
		TotalValueFont: 10.0 * scale,
		TotalHeight:    8.5 * scale,

		WordsPadY:      1.6 * scale,
		ClosingFont:    8.5 * scale,
		ClosingHeight:  4.4 * scale,
		ClosingPadY:    1.8 * scale,
		SignLabelGap:   2.0 * scale,
		SignNameFont:   9.0 * scale,
		SignNameHeight: 4.6 * scale,
		SignRuleGap:    1.6 * scale,
		SignRuleWidth:  42.0 * scale,
		SignCaption:    7.0 * scale,
		SignCaptionGap: 1.4 * scale,
		SignCaptionH:   3.2 * scale,
		SignPadY:       2.4 * scale,

		ColNo:    colNo,
		ColName:  colName,
		ColQty:   colQty,
		ColPrice: colPrice,
		ColSum:   colSum,
		ColLabel: contentWidth * 0.2800,

		Border: 0.2,
	}
}

// EN: Function `layoutTabularAct`.
//
// EN: What it does: layoutTabularAct walks blank №5 once and either measures it or draws it.
//
// EN: Key points: every element is a row of the outer frame, so the walk simply stacks row heights and finally
// EN: closes the frame around them.
func layoutTabularAct(pdf *gofpdf.Fpdf, data actPDFData, layout tabularActLayout, render bool) float64 {
	left := layout.Left
	width := layout.ContentWidth
	right := left + width
	top := layout.Top
	y := top

	if render {
		pdf.SetLineWidth(layout.Border)
	}

	// Title cell.
	titleHeight := layout.TitlePadY + layout.TitleHeight + layout.SubtitleHeight + layout.TitlePadY
	if render {
		pdf.SetFont("Mercel", "B", layout.TitleFont)
		title := "АКТ ПРИЁМА-СДАЧИ ОКАЗАННЫХ УСЛУГ"
		titleWidth := trackedTextWidth(pdf, title, layout.TitleTracking)
		drawTrackedText(pdf, left+((width-titleWidth)/2), y+layout.TitlePadY, layout.TitleHeight, title, layout.TitleTracking)
		pdf.SetFont("Mercel", "", layout.SubtitleFont)
		pdf.SetXY(left, y+layout.TitlePadY+layout.TitleHeight)
		pdf.CellFormat(width, layout.SubtitleHeight, tabularActSubtitle(data), "", 0, "C", false, 0, "")
	}
	y = drawTabularRule(pdf, left, right, y+titleHeight, render)

	// Labelled information rows.
	for _, row := range tabularActInfoRows(data) {
		y = drawTabularInfoRow(pdf, layout, y, row[0], row[1], render)
	}

	// Price list header.
	if render {
		pdf.SetFont("Mercel", "B", layout.HeaderFont)
		headerY := y + ((layout.HeaderHeight - layout.LabelHeight) / 2)
		for _, x := range []float64{left + layout.ColNo, left + layout.ColNo + layout.ColName, left + layout.ColNo + layout.ColName + layout.ColQty, right - layout.ColSum} {
			pdf.Line(x, y, x, y+layout.HeaderHeight)
		}
		drawTabularHeaderCell(pdf, layout, left, layout.ColNo, headerY, "№", "C")
		drawTabularHeaderCell(pdf, layout, left+layout.ColNo, layout.ColName, headerY, "НАИМЕНОВАНИЕ УСЛУГИ", "L")
		drawTabularHeaderCell(pdf, layout, left+layout.ColNo+layout.ColName, layout.ColQty, headerY, "КОЛ-ВО", "C")
		drawTabularHeaderCell(pdf, layout, left+layout.ColNo+layout.ColName+layout.ColQty, layout.ColPrice, headerY, "ЦЕНА", "R")
		drawTabularHeaderCell(pdf, layout, right-layout.ColSum, layout.ColSum, headerY, "СУММА, РУБ.", "R")
	}
	y = drawTabularRule(pdf, left, right, y+layout.HeaderHeight, render)

	// Price list rows.
	y = drawTabularItems(pdf, layout, data, y, render)

	// Total row.
	if render {
		pdf.Line(right-layout.ColSum, y, right-layout.ColSum, y+layout.TotalHeight)
		pdf.SetFont("Mercel", "B", layout.TotalFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(width-layout.ColSum-layout.CellPadX, layout.TotalHeight, "ИТОГО К ОПЛАТЕ", "", 0, "R", false, 0, "")
		pdf.SetFont("Mercel", "B", layout.TotalValueFont)
		pdf.SetXY(right-layout.ColSum, y)
		pdf.CellFormat(layout.ColSum-layout.CellPadX, layout.TotalHeight, formatThousands(data.TotalAmount), "", 0, "R", false, 0, "")
	}
	y = drawTabularRule(pdf, left, right, y+layout.TotalHeight, render)

	// Amount in words.
	y = drawTabularInfoRow(pdf, layout, y, "СУММА ПРОПИСЬЮ", tabularActWords(data), render)

	// Closing statement.
	pdf.SetFont("Mercel", "", layout.ClosingFont)
	closingWidth := width - (layout.CellPadX * 2)
	closingLines := pdf.SplitText(typoClosingText(), closingWidth)
	if len(closingLines) == 0 {
		closingLines = []string{""}
	}
	if render {
		for index, line := range closingLines {
			pdf.SetXY(left+layout.CellPadX, y+layout.ClosingPadY+(float64(index)*layout.ClosingHeight))
			pdf.CellFormat(closingWidth, layout.ClosingHeight, line, "", 0, "L", false, 0, "")
		}
	}
	y = drawTabularRule(pdf, left, right, y+(layout.ClosingPadY*2)+(float64(len(closingLines))*layout.ClosingHeight), render)

	// Signatures.
	y = drawTabularSignatures(pdf, layout, data, y, render)

	if render {
		// Close the frame: outer box plus the vertical rule between the signature columns.
		pdf.SetLineWidth(layout.Border)
		pdf.Rect(left, top, width, y-top, "")
	}
	return y
}

// EN: Function `drawTabularRule`.
//
// EN: What it does: drawTabularRule closes one row of the frame with a horizontal line.
//
// EN: Key points: returns the same y it was given so the caller can keep stacking rows.
func drawTabularRule(pdf *gofpdf.Fpdf, left float64, right float64, y float64, render bool) float64 {
	if render {
		pdf.Line(left, y, right, y)
	}
	return y
}

// EN: Function `drawTabularHeaderCell`.
//
// EN: What it does: drawTabularHeaderCell prints one tracked heading of the price list.
func drawTabularHeaderCell(pdf *gofpdf.Fpdf, layout tabularActLayout, x float64, width float64, y float64, text string, align string) {
	switch align {
	case "R":
		drawTypoTrackedRight(pdf, x+width-layout.CellPadX, y, layout.LabelHeight, text, layout.HeaderTracking)
	case "C":
		drawTrackedText(pdf, x+((width-trackedTextWidth(pdf, text, layout.HeaderTracking))/2), y, layout.LabelHeight, text, layout.HeaderTracking)
	default:
		drawTrackedText(pdf, x+layout.CellPadX, y, layout.LabelHeight, text, layout.HeaderTracking)
	}
}

// EN: Function `drawTabularInfoRow`.
//
// EN: What it does: drawTabularInfoRow renders one "label | value" row of the frame.
//
// EN: Key points: the value may wrap, so the row height comes from the measured value text.
func drawTabularInfoRow(pdf *gofpdf.Fpdf, layout tabularActLayout, y float64, label string, value string, render bool) float64 {
	left := layout.Left
	right := left + layout.ContentWidth
	valueX := left + layout.ColLabel
	valueWidth := layout.ContentWidth - layout.ColLabel - layout.CellPadX

	pdf.SetFont("Mercel", "", layout.ValueFont)
	lines := pdf.SplitText(value, valueWidth-layout.CellPadX)
	if len(lines) == 0 {
		lines = []string{""}
	}
	height := layout.InfoPadY + (float64(len(lines)) * layout.ValueLineHeight) + layout.InfoPadY

	if render {
		drawTabularLabel(pdf, layout, left+layout.CellPadX, y+layout.InfoPadY+((layout.ValueLineHeight-layout.LabelHeight)/2), layout.ColLabel-(layout.CellPadX*2), label)
		pdf.SetFont("Mercel", "", layout.ValueFont)
		for index, line := range lines {
			pdf.SetXY(valueX, y+layout.InfoPadY+(float64(index)*layout.ValueLineHeight))
			pdf.CellFormat(valueWidth, layout.ValueLineHeight, line, "", 0, "L", false, 0, "")
		}
		pdf.Line(valueX-layout.CellPadX, y, valueX-layout.CellPadX, y+height)
	}
	return drawTabularRule(pdf, left, right, y+height, render)
}

// EN: Function `drawTabularItems`.
//
// EN: What it does: drawTabularItems renders the price list rows of blank №5 inside the frame.
//
// EN: Key points: vertical rules are drawn per row so the grid stays continuous whatever the row heights are.
func drawTabularItems(pdf *gofpdf.Fpdf, layout tabularActLayout, data actPDFData, y float64, render bool) float64 {
	left := layout.Left
	right := left + layout.ContentWidth
	padX := layout.CellPadX

	xName := left + layout.ColNo
	xQty := xName + layout.ColName
	xPrice := xQty + layout.ColQty
	xSum := xPrice + layout.ColPrice

	pdf.SetFont("Mercel", "", layout.BodyFont)
	for index, item := range data.Items {
		lines := pdf.SplitText(item.Name, layout.ColName-(padX*2))
		if len(lines) == 0 {
			lines = []string{""}
		}
		rowHeight := (float64(len(lines)) * layout.BodyLineHeight) + (layout.RowPadY * 2)
		if rowHeight < layout.RowMinHeight {
			rowHeight = layout.RowMinHeight
		}
		if render {
			pdf.SetFont("Mercel", "", layout.BodyFont)
			for _, x := range []float64{xName, xQty, xPrice, xSum} {
				pdf.Line(x, y, x, y+rowHeight)
			}

			pdf.SetXY(left, y)
			pdf.CellFormat(layout.ColNo, rowHeight, fmt.Sprintf("%d", index+1), "", 0, "C", false, 0, "")

			textY := y + ((rowHeight - (float64(len(lines)) * layout.BodyLineHeight)) / 2)
			for lineIndex, line := range lines {
				pdf.SetXY(xName+padX, textY+(float64(lineIndex)*layout.BodyLineHeight))
				pdf.CellFormat(layout.ColName-(padX*2), layout.BodyLineHeight, line, "", 0, "L", false, 0, "")
			}

			pdf.SetXY(xQty, y)
			pdf.CellFormat(layout.ColQty, rowHeight, formatDocumentQuantity(item.Quantity, item.Unit), "", 0, "C", false, 0, "")
			pdf.SetXY(xPrice, y)
			pdf.CellFormat(layout.ColPrice-padX, rowHeight, formatThousands(item.Rate), "", 0, "R", false, 0, "")
			pdf.SetXY(xSum, y)
			pdf.CellFormat(layout.ColSum-padX, rowHeight, formatThousands(item.LineTotal), "", 0, "R", false, 0, "")
		}
		y = drawTabularRule(pdf, left, right, y+rowHeight, render)
	}
	return y
}

// EN: Function `drawTabularSignatures`.
//
// EN: What it does: drawTabularSignatures renders the closing signature row of blank №5.
//
// EN: Key points: both columns share one framed row, split by a vertical rule.
func drawTabularSignatures(pdf *gofpdf.Fpdf, layout tabularActLayout, data actPDFData, y float64, render bool) float64 {
	left := layout.Left
	right := left + layout.ContentWidth
	middle := left + (layout.ContentWidth * 0.4088)

	height := layout.SignPadY + layout.LabelHeight + layout.SignLabelGap + layout.SignNameHeight +
		layout.SignRuleGap + layout.SignCaptionGap + layout.SignCaptionH + layout.SignPadY

	if render {
		pdf.Line(middle, y, middle, y+height)
		drawTabularSignatureColumn(pdf, layout, left+layout.CellPadX, y, "ИСПОЛНИТЕЛЬ", shortEmployeeSignatureName(data.EmployeeFullName))
		drawTabularSignatureColumn(pdf, layout, middle+layout.CellPadX, y, "ЗАКАЗЧИК", resolveContractDirectorShort(data.ContractTitle))
	}
	return drawTabularRule(pdf, left, right, y+height, render)
}

// EN: Function `drawTabularSignatureColumn`.
//
// EN: What it does: drawTabularSignatureColumn stacks the party label, the printed name, the rule and its caption.
func drawTabularSignatureColumn(pdf *gofpdf.Fpdf, layout tabularActLayout, x float64, y float64, label string, name string) {
	cursor := y + layout.SignPadY
	drawTabularLabel(pdf, layout, x, cursor, layout.SignRuleWidth, label)
	cursor += layout.LabelHeight + layout.SignLabelGap

	pdf.SetFont("Mercel", "", layout.SignNameFont)
	pdf.SetXY(x, cursor)
	pdf.CellFormat(layout.SignRuleWidth, layout.SignNameHeight, name, "", 0, "L", false, 0, "")
	cursor += layout.SignNameHeight + layout.SignRuleGap

	pdf.Line(x, cursor, x+layout.SignRuleWidth, cursor)
	cursor += layout.SignCaptionGap

	pdf.SetFont("Mercel", "", layout.SignCaption)
	pdf.SetXY(x, cursor)
	pdf.CellFormat(layout.SignRuleWidth, layout.SignCaptionH, "подпись / расшифровка", "", 0, "L", false, 0, "")
}

// EN: Function `drawTabularLabel`.
//
// EN: What it does: drawTabularLabel prints a tracked caps label and shrinks it until it fits its column.
//
// EN: Key points: the labels of this blank differ a lot in length ("ОСНОВАНИЕ" against "ПЕРИОД ОКАЗАНИЯ УСЛУГ"),
// EN: so a fixed size would let the longest one run into the value column.
func drawTabularLabel(pdf *gofpdf.Fpdf, layout tabularActLayout, x float64, y float64, width float64, text string) {
	size := layout.LabelFont
	for size > 3.0 {
		pdf.SetFont("Mercel", "B", size)
		if trackedTextWidth(pdf, text, layout.LabelTracking) <= width {
			break
		}
		size -= 0.2
	}
	drawTrackedText(pdf, x, y, layout.LabelHeight, text, layout.LabelTracking)
}

// EN: Function `tabularActSubtitle`.
//
// EN: What it does: tabularActSubtitle builds the number/date/city line under the title of blank №5.
func tabularActSubtitle(data actPDFData) string {
	return fmt.Sprintf("№ %d    от %s    г. Санкт-Петербург", data.ActNumber, russianActDate(data.GeneratedAt))
}

// EN: Function `tabularActInfoRows`.
//
// EN: What it does: tabularActInfoRows builds the labelled rows describing the parties and the basis.
//
// EN: Key points: the service period is not stored with the calculation, so it is taken as the calendar month the
// EN: act is dated with, which is how these acts are issued.
func tabularActInfoRows(data actPDFData) [][2]string {
	return [][2]string{
		{"ЗАКАЗЧИК", fmt.Sprintf("ООО «%s», в лице генерального директора %s, действующего на основании Устава", data.CustomerName, resolveContractDirectorName(data.ContractTitle))},
		{"ИСПОЛНИТЕЛЬ", fmt.Sprintf("Индивидуальный предприниматель %s", strings.TrimSpace(data.EmployeeFullName))},
		{"ОСНОВАНИЕ", fmt.Sprintf("Договор № %s от %s", data.ContractNumber, formalContractDate(data.ContractDate))},
		{"ПЕРИОД ОКАЗАНИЯ УСЛУГ", tabularActPeriod(data.GeneratedAt)},
	}
}

// EN: Function `tabularActPeriod`.
//
// EN: What it does: tabularActPeriod returns the calendar month of the act as a "с … по …" period.
func tabularActPeriod(value time.Time) string {
	first := time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, value.Location())
	last := first.AddDate(0, 1, -1)
	return fmt.Sprintf("с %s по %s", first.Format("02.01.2006"), last.Format("02.01.2006"))
}

// EN: Function `tabularActWords`.
//
// EN: What it does: tabularActWords spells out the total for the "сумма прописью" row of blank №5.
func tabularActWords(data actPDFData) string {
	words := []rune(fmt.Sprintf("%s %s 00 копеек, без НДС.", data.TotalWords, data.TotalCurrency))
	if len(words) == 0 {
		return ""
	}
	return strings.ToUpper(string(words[0])) + string(words[1:])
}
