package appcore

import (
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// EN: Typewritten act blank (Акт_бланк_10).
//
// EN: What it does: this file renders the typewriter flavoured variant — everything is set in Courier New, the
// EN: preamble sits in a heavy framed box and the price list is a plain fully gridded table.
//
// EN: Key points: the monospaced face is what tells it apart from the other blanks at a glance; the bold weight is
// EN: deliberate — regular Courier is a very light face and prints grey rather than black at this size, while the
// EN: bold one reads like an actual typewritten sheet. Like every blank it is scaled down until it fits a single A4
// EN: sheet, and one walk both measures and draws it.
const (
	typewriterActTargetScale = 1.00
	typewriterActMinScale    = 0.60
)

type typewriterActLayout struct {
	Scale        float64
	Left         float64
	Top          float64
	Bottom       float64
	ContentWidth float64

	BoxPad        float64
	BoxBorder     float64
	TitleFont     float64
	TitleTracking float64
	TitleHeight   float64
	TitleGap      float64

	SubtitleFont   float64
	SubtitleHeight float64
	SubtitleGap    float64

	MetaFont   float64
	MetaHeight float64
	MetaGap    float64

	PartyHeight float64
	PartyGap    float64
	PartyColumn float64

	IntroLineHeight float64
	BoxGap          float64

	TableHeaderFont   float64
	TableHeaderHeight float64
	TableBodyFont     float64
	TableLineHeight   float64
	TableRowPadY      float64
	TableCellPadX     float64
	TableRowMinHeight float64
	TableTotalHeight  float64
	TableGap          float64
	Border            float64

	BodyFont       float64
	BodyLineHeight float64
	WordsGap       float64
	ClosingGap     float64
	ClosingLineGap float64

	SignLabelHeight float64
	SignLabelGap    float64
	SignLineHeight  float64
	SignGap         float64
	SignColumn      float64

	ColNo    float64
	ColName  float64
	ColQty   float64
	ColPrice float64
	ColSum   float64
}

// EN: Function `renderTypewriterAct`.
//
// EN: What it does: renderTypewriterAct fits blank №10 to one page and draws it, returning the bottom edge.
//
// EN: Key points: mirrors the other renderers so the shared page checks in renderActPDF apply here too.
func renderTypewriterAct(pdf *gofpdf.Fpdf, data actPDFData) (float64, error) {
	layout, err := fitTypewriterActLayout(pdf, data)
	if err != nil {
		return 0, err
	}
	layoutTypewriterAct(pdf, data, layout, true)
	return layout.Bottom, nil
}

// EN: Function `fitTypewriterActLayout`.
//
// EN: What it does: fitTypewriterActLayout picks the largest scale that keeps blank №10 on a single sheet.
//
// EN: Key points: Courier is a wide face, so this blank needs the most head room of all five before the export is
// EN: refused with the shared hint.
func fitTypewriterActLayout(pdf *gofpdf.Fpdf, data actPDFData) (typewriterActLayout, error) {
	scale, ok := fitActScale(typewriterActMinScale, typewriterActTargetScale, func(scale float64) bool {
		candidate := newTypewriterActLayout(pdf, scale)
		return layoutTypewriterAct(pdf, data, candidate, false) <= candidate.Bottom
	})
	if !ok {
		return typewriterActLayout{}, errActDoesNotFit()
	}
	return newTypewriterActLayout(pdf, scale), nil
}

// EN: Function `newTypewriterActLayout`.
//
// EN: What it does: newTypewriterActLayout converts the Word measurements of blank №10 into millimetres for one scale.
//
// EN: Key points: sizes come from the source document at scale 1.0 (8.5 pt body, 20 mm side margins, a 5.3 mm
// EN: padded box); margins shrink with the content but never below the ones used by the standard blank.
func newTypewriterActLayout(pdf *gofpdf.Fpdf, scale float64) typewriterActLayout {
	pageWidth, pageHeight := pdf.GetPageSize()

	left := 20.0 * scale
	if left < 12.0 {
		left = 12.0
	}
	top := 15.9 * scale
	if top < 11.0 {
		top = 11.0
	}
	bottomMargin := 15.9 * scale
	if bottomMargin < 10.0 {
		bottomMargin = 10.0
	}
	contentWidth := pageWidth - (left * 2)

	// Column proportions of the source table: 500 / 3938 / 1200 / 1500 / 1900 twips.
	colNo := contentWidth * 0.0553
	colQty := contentWidth * 0.1328
	colPrice := contentWidth * 0.1660
	colSum := contentWidth * 0.2102
	colName := contentWidth - colNo - colQty - colPrice - colSum

	return typewriterActLayout{
		Scale:        scale,
		Left:         left,
		Top:          top,
		Bottom:       pageHeight - bottomMargin,
		ContentWidth: contentWidth,

		BoxPad:        5.3 * scale,
		BoxBorder:     0.53,
		TitleFont:     12.0 * scale,
		TitleTracking: 0.7 * scale,
		TitleHeight:   6.2 * scale,
		TitleGap:      1.6 * scale,

		SubtitleFont:   8.5 * scale,
		SubtitleHeight: 4.4 * scale,
		SubtitleGap:    4.6 * scale,

		MetaFont:   8.5 * scale,
		MetaHeight: 4.4 * scale,
		MetaGap:    5.3 * scale,

		PartyHeight: 4.4 * scale,
		PartyGap:    2.3 * scale,
		PartyColumn: 40.6 * scale,

		IntroLineHeight: 4.6 * scale,
		BoxGap:          5.3 * scale,

		TableHeaderFont:   8.0 * scale,
		TableHeaderHeight: 6.6 * scale,
		TableBodyFont:     8.5 * scale,
		TableLineHeight:   4.2 * scale,
		TableRowPadY:      1.2 * scale,
		TableCellPadX:     1.9 * scale,
		TableRowMinHeight: 7.1 * scale,
		TableTotalHeight:  7.4 * scale,
		TableGap:          5.3 * scale,
		Border:            0.2,

		BodyFont:       8.5 * scale,
		BodyLineHeight: 4.9 * scale,
		WordsGap:       5.3 * scale,
		ClosingGap:     8.8 * scale,
		ClosingLineGap: 2.3 * scale,

		SignLabelHeight: 4.4 * scale,
		SignLabelGap:    2.8 * scale,
		SignLineHeight:  4.4 * scale,
		SignGap:         1.1 * scale,
		SignColumn:      contentWidth * 0.4877,

		ColNo:    colNo,
		ColName:  colName,
		ColQty:   colQty,
		ColPrice: colPrice,
		ColSum:   colSum,
	}
}

// EN: Function `layoutTypewriterAct`.
//
// EN: What it does: layoutTypewriterAct walks blank №10 once and either measures it or draws it.
//
// EN: Key points: measuring and drawing share this walk, so the fitted scale always matches what ends up on the
// EN: page; the function returns the y coordinate right below the last element.
func layoutTypewriterAct(pdf *gofpdf.Fpdf, data actPDFData, layout typewriterActLayout, render bool) float64 {
	left := layout.Left
	width := layout.ContentWidth
	y := layout.Top

	y = drawTypewriterHeaderBox(pdf, layout, data, y, render) + layout.BoxGap
	y = drawTypewriterTable(pdf, layout, data, y, render) + layout.TableGap

	y = drawActParagraph(pdf, left, width, y, typewriterActWords(data), "B", layout.BodyFont, layout.BodyLineHeight, "L", render) + layout.WordsGap
	y = drawActParagraph(pdf, left, width, y, "Услуги выполнены полностью и в срок.", "B", layout.BodyFont, layout.BodyLineHeight, "L", render) + layout.ClosingLineGap
	y = drawActParagraph(pdf, left, width, y, "Заказчик претензий по объёму, качеству и срокам не имеет.", "B", layout.BodyFont, layout.BodyLineHeight, "L", render) + layout.ClosingGap

	y = drawTypewriterSignatures(pdf, layout, data, y, render)
	return y
}

// EN: Function `drawTypewriterHeaderBox`.
//
// EN: What it does: drawTypewriterHeaderBox renders the framed preamble: title, parties and the basis sentence.
//
// EN: Key points: the box is drawn last, once its height is known, so the frame always wraps the text exactly.
func drawTypewriterHeaderBox(pdf *gofpdf.Fpdf, layout typewriterActLayout, data actPDFData, y float64, render bool) float64 {
	boxLeft := layout.Left
	boxWidth := layout.ContentWidth
	innerLeft := boxLeft + layout.BoxPad
	innerWidth := boxWidth - (layout.BoxPad * 2)
	cursor := y + layout.BoxPad

	if render {
		pdf.SetFont("Mercel", "B", layout.TitleFont)
		title := "А К Т"
		titleWidth := trackedTextWidth(pdf, title, layout.TitleTracking)
		drawTrackedText(pdf, innerLeft+((innerWidth-titleWidth)/2), cursor, layout.TitleHeight, title, layout.TitleTracking)
	}
	cursor += layout.TitleHeight + layout.TitleGap

	if render {
		pdf.SetFont("Mercel", "B", layout.SubtitleFont)
		pdf.SetXY(innerLeft, cursor)
		pdf.CellFormat(innerWidth, layout.SubtitleHeight, "приёма-сдачи оказанных услуг", "", 0, "C", false, 0, "")
	}
	cursor += layout.SubtitleHeight + layout.SubtitleGap

	if render {
		pdf.SetFont("Mercel", "B", layout.MetaFont)
		pdf.SetXY(innerLeft, cursor)
		pdf.CellFormat(innerWidth/2, layout.MetaHeight, fmt.Sprintf("N %d   г. Санкт-Петербург", data.ActNumber), "", 0, "L", false, 0, "")
		pdf.CellFormat(innerWidth/2, layout.MetaHeight, russianActDate(data.GeneratedAt), "", 0, "R", false, 0, "")
	}
	cursor += layout.MetaHeight + layout.MetaGap

	parties := [][2]string{
		{"ЗАКАЗЧИК", fmt.Sprintf("ООО «%s»", data.CustomerName)},
		{"ИСПОЛНИТЕЛЬ", fmt.Sprintf("ИП %s", strings.TrimSpace(data.EmployeeFullName))},
	}
	for _, party := range parties {
		if render {
			pdf.SetFont("Mercel", "B", layout.MetaFont)
			pdf.SetXY(innerLeft, cursor)
			pdf.CellFormat(layout.PartyColumn, layout.PartyHeight, party[0], "", 0, "L", false, 0, "")
			pdf.SetXY(innerLeft+layout.PartyColumn, cursor)
			pdf.CellFormat(innerWidth-layout.PartyColumn, layout.PartyHeight, party[1], "", 0, "L", false, 0, "")
		}
		cursor += layout.PartyHeight + layout.PartyGap
	}
	cursor += layout.MetaGap - layout.PartyGap

	cursor = drawActParagraph(pdf, innerLeft, innerWidth, cursor, typewriterActIntro(data), "B", layout.MetaFont, layout.IntroLineHeight, "L", render)
	cursor += layout.BoxPad

	if render {
		pdf.SetLineWidth(layout.BoxBorder)
		pdf.Rect(boxLeft, y, boxWidth, cursor-y, "")
	}
	return cursor
}

// EN: Function `drawTypewriterTable`.
//
// EN: What it does: drawTypewriterTable renders the fully gridded five column price list of blank №10.
//
// EN: Key points: every cell is boxed, and the total row merges the first three columns exactly like the source.
func drawTypewriterTable(pdf *gofpdf.Fpdf, layout typewriterActLayout, data actPDFData, y float64, render bool) float64 {
	left := layout.Left
	padX := layout.TableCellPadX

	xName := left + layout.ColNo
	xQty := xName + layout.ColName
	xPrice := xQty + layout.ColQty
	xSum := xPrice + layout.ColPrice

	if render {
		pdf.SetLineWidth(layout.Border)
		pdf.SetFont("Mercel", "B", layout.TableHeaderFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(layout.ColNo, layout.TableHeaderHeight, "№", "1", 0, "C", false, 0, "")
		pdf.CellFormat(layout.ColName, layout.TableHeaderHeight, " НАИМЕНОВАНИЕ УСЛУГИ", "1", 0, "L", false, 0, "")
		pdf.CellFormat(layout.ColQty, layout.TableHeaderHeight, "КОЛ-ВО", "1", 0, "C", false, 0, "")
		pdf.CellFormat(layout.ColPrice, layout.TableHeaderHeight, "ЦЕНА ", "1", 0, "R", false, 0, "")
		pdf.CellFormat(layout.ColSum, layout.TableHeaderHeight, "СУММА ", "1", 0, "R", false, 0, "")
	}
	y += layout.TableHeaderHeight

	pdf.SetFont("Mercel", "B", layout.TableBodyFont)
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
			pdf.SetFont("Mercel", "B", layout.TableBodyFont)
			pdf.Rect(left, y, layout.ColNo, rowHeight, "")
			pdf.Rect(xName, y, layout.ColName, rowHeight, "")
			pdf.Rect(xQty, y, layout.ColQty, rowHeight, "")
			pdf.Rect(xPrice, y, layout.ColPrice, rowHeight, "")
			pdf.Rect(xSum, y, layout.ColSum, rowHeight, "")

			pdf.SetXY(left, y)
			pdf.CellFormat(layout.ColNo, rowHeight, fmt.Sprintf("%d", index+1), "", 0, "C", false, 0, "")

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
	}

	if render {
		pdf.SetFont("Mercel", "B", layout.TableBodyFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(layout.ColNo+layout.ColName+layout.ColQty, layout.TableTotalHeight, "", "1", 0, "L", false, 0, "")
		pdf.CellFormat(layout.ColPrice, layout.TableTotalHeight, "ИТОГО: ", "1", 0, "R", false, 0, "")
		pdf.CellFormat(layout.ColSum, layout.TableTotalHeight, formatThousands(data.TotalAmount)+" ", "1", 0, "R", false, 0, "")
	}
	return y + layout.TableTotalHeight
}

// EN: Function `drawTypewriterSignatures`.
//
// EN: What it does: drawTypewriterSignatures renders the two signature columns closing blank №10.
//
// EN: Key points: the signature line keeps the underscores of the source blank and carries the printed name in
// EN: slashes, which is how a typewritten act is signed.
func drawTypewriterSignatures(pdf *gofpdf.Fpdf, layout typewriterActLayout, data actPDFData, y float64, render bool) float64 {
	left := layout.Left
	rightColumn := left + layout.SignColumn
	columnWidth := layout.ContentWidth - layout.SignColumn

	if render {
		pdf.SetFont("Mercel", "B", layout.BodyFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(layout.SignColumn, layout.SignLabelHeight, "ИСПОЛНИТЕЛЬ:", "", 0, "L", false, 0, "")
		pdf.SetXY(rightColumn, y)
		pdf.CellFormat(columnWidth, layout.SignLabelHeight, "ЗАКАЗЧИК:", "", 0, "L", false, 0, "")
	}
	y += layout.SignLabelHeight + layout.SignLabelGap

	if render {
		pdf.SetFont("Mercel", "B", layout.BodyFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(layout.SignColumn, layout.SignLineHeight, typewriterSignatureLine(shortEmployeeSignatureName(data.EmployeeFullName)), "", 0, "L", false, 0, "")
		pdf.SetXY(rightColumn, y)
		pdf.CellFormat(columnWidth, layout.SignLineHeight, typewriterSignatureLine(resolveContractDirectorShort(data.ContractTitle)), "", 0, "L", false, 0, "")
	}
	return y + layout.SignLineHeight + layout.SignGap
}

// EN: Function `typewriterSignatureLine`.
//
// EN: What it does: typewriterSignatureLine builds the "signature line / name" pair of blank №10.
func typewriterSignatureLine(name string) string {
	return fmt.Sprintf("______________________ /%s/", strings.TrimSpace(name))
}

// EN: Function `typewriterActIntro`.
//
// EN: What it does: typewriterActIntro builds the basis sentence closing the framed preamble.
//
// EN: Key points: keeps the wording of the source blank, which refers to the contract with a plain "N" the way a
// EN: typewriter would.
func typewriterActIntro(data actPDFData) string {
	return fmt.Sprintf(
		"Стороны составили настоящий акт согласно Договору N %s от %s, заключённому между Сторонами, о том, что Исполнитель выполнил, а Заказчик принял следующие работы:",
		data.ContractNumber,
		formalContractDate(data.ContractDate),
	)
}

// EN: Function `typewriterActWords`.
//
// EN: What it does: typewriterActWords spells out the total for blank №10.
func typewriterActWords(data actPDFData) string {
	return fmt.Sprintf("СУММА ПРОПИСЬЮ: %s %s 00 копеек, без НДС.", data.TotalWords, data.TotalCurrency)
}
