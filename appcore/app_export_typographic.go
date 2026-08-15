package appcore

import (
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// EN: Typographic act blank (Акт_бланк_3).
//
// EN: What it does: this file renders the second variant of the act — a flush-left typographic layout with a
// EN: letter-spaced title, section labels, a five column price list and rule-only table borders.
//
// EN: Key points: the source Word blank spills over two pages, so the whole document is scaled down until it fits a
// EN: single A4 sheet; measuring and drawing share one walk over the document, which keeps the fit exact.
const (
	typoActTargetScale = 1.00
	typoActMinScale    = 0.66
)

type typoActLayout struct {
	Scale        float64
	Left         float64
	Top          float64
	Bottom       float64
	ContentWidth float64

	TitleFont     float64
	TitleTracking float64
	TitleHeight   float64
	TitleGap      float64

	SubtitleFont   float64
	SubtitleHeight float64
	SubtitleGap    float64
	RuleGap        float64

	MetaFont   float64
	MetaHeight float64
	MetaGap    float64

	SectionFont     float64
	SectionTracking float64
	SectionHeight   float64
	SectionGap      float64

	BodyFont       float64
	BodyLineHeight float64
	BodyGap        float64
	BasisGap       float64

	TableHeaderFont     float64
	TableHeaderTracking float64
	TableHeaderHeight   float64
	TableBodyFont       float64
	TableLineHeight     float64
	TableRowPadY        float64
	TableCellPadX       float64
	TableRowMinHeight   float64
	TableTotalHeight    float64
	TableTotalFont      float64
	TableTotalValueFont float64
	TableGap            float64

	TotalsFont       float64
	TotalsLineHeight float64
	TotalsGap        float64

	ClosingFont       float64
	ClosingLineHeight float64
	ClosingGap        float64

	SignLabelFont     float64
	SignLabelTracking float64
	SignLabelHeight   float64
	SignPadTop        float64
	SignNameFont      float64
	SignNameHeight    float64
	SignNameGap       float64
	SignRuleGap       float64
	SignRuleWidth     float64
	SignCaptionFont   float64
	SignCaptionGap    float64
	SignCaptionHeight float64
	SignGutter        float64

	ColNo    float64
	ColName  float64
	ColQty   float64
	ColPrice float64
	ColSum   float64

	RuleBold float64
	RuleThin float64
	RuleHair float64
}

// EN: Function `renderTypographicAct`.
//
// EN: What it does: renderTypographicAct fits blank №3 to one page and draws it, returning the bottom edge.
//
// EN: Key points: mirrors renderClassicAct so the shared page checks in renderActPDF apply to both blanks.
func renderTypographicAct(pdf *gofpdf.Fpdf, data actPDFData) (float64, error) {
	layout, err := fitTypoActLayout(pdf, data)
	if err != nil {
		return 0, err
	}
	layoutTypographicAct(pdf, data, layout, true)
	return layout.Bottom, nil
}

// EN: Function `fitTypoActLayout`.
//
// EN: What it does: fitTypoActLayout picks the largest scale that keeps the typographic blank on a single sheet.
//
// EN: Key points: the Word original overflows onto a second page, so shrinking is expected for long price lists;
// EN: when even the smallest scale overflows the export is refused with the usual hint.
func fitTypoActLayout(pdf *gofpdf.Fpdf, data actPDFData) (typoActLayout, error) {
	scale, ok := fitActScale(typoActMinScale, typoActTargetScale, func(scale float64) bool {
		candidate := newTypoActLayout(pdf, scale)
		return layoutTypographicAct(pdf, data, candidate, false) <= candidate.Bottom
	})
	if !ok {
		return typoActLayout{}, errActDoesNotFit()
	}
	return newTypoActLayout(pdf, scale), nil
}

// EN: Function `newTypoActLayout`.
//
// EN: What it does: newTypoActLayout converts the Word measurements of blank №3 into millimetres for one scale.
//
// EN: Key points: font sizes, spacing and rules come from the source document at scale 1.0; page margins shrink
// EN: together with the content but never below the margins used by the classic blank.
func newTypoActLayout(pdf *gofpdf.Fpdf, scale float64) typoActLayout {
	pageWidth, pageHeight := pdf.GetPageSize()

	left := 22.9 * scale
	if left < 12.0 {
		left = 12.0
	}
	top := 21.2 * scale
	if top < 12.0 {
		top = 12.0
	}
	bottomMargin := 20.0 * scale
	if bottomMargin < 10.0 {
		bottomMargin = 10.0
	}
	contentWidth := pageWidth - (left * 2)

	colNo := 9.0
	colQty := 20.0
	colPrice := 26.0
	colSum := 30.0
	colName := contentWidth - colNo - colQty - colPrice - colSum

	return typoActLayout{
		Scale:        scale,
		Left:         left,
		Top:          top,
		Bottom:       pageHeight - bottomMargin,
		ContentWidth: contentWidth,

		TitleFont:     22.0 * scale,
		TitleTracking: 1.6 * scale,
		TitleHeight:   9.6 * scale,
		TitleGap:      1.1 * scale,

		SubtitleFont:   11.0 * scale,
		SubtitleHeight: 5.0 * scale,
		SubtitleGap:    2.8 * scale,
		RuleGap:        1.4 * scale,

		MetaFont:   9.5 * scale,
		MetaHeight: 4.6 * scale,
		MetaGap:    7.4 * scale,

		SectionFont:     8.0 * scale,
		SectionTracking: 0.35 * scale,
		SectionHeight:   3.8 * scale,
		SectionGap:      2.5 * scale,

		BodyFont:       10.5 * scale,
		BodyLineHeight: 5.3 * scale,
		BodyGap:        3.5 * scale,
		BasisGap:       7.4 * scale,

		TableHeaderFont:     8.0 * scale,
		TableHeaderTracking: 0.25 * scale,
		TableHeaderHeight:   6.6 * scale,
		TableBodyFont:       9.5 * scale,
		TableLineHeight:     4.6 * scale,
		TableRowPadY:        1.4 * scale,
		TableCellPadX:       1.9 * scale,
		TableRowMinHeight:   7.9 * scale,
		TableTotalHeight:    9.2 * scale,
		TableTotalFont:      9.0 * scale,
		TableTotalValueFont: 11.0 * scale,
		TableGap:            6.0 * scale,

		TotalsFont:       10.5 * scale,
		TotalsLineHeight: 5.3 * scale,
		TotalsGap:        6.0 * scale,

		ClosingFont:       10.5 * scale,
		ClosingLineHeight: 5.3 * scale,
		ClosingGap:        12.3 * scale,

		SignLabelFont:     8.0 * scale,
		SignLabelTracking: 0.25 * scale,
		SignLabelHeight:   3.8 * scale,
		SignPadTop:        2.8 * scale,
		SignNameFont:      9.5 * scale,
		SignNameHeight:    4.6 * scale,
		SignNameGap:       2.4 * scale,
		SignRuleGap:       1.6 * scale,
		SignRuleWidth:     46.0 * scale,
		SignCaptionFont:   7.0 * scale,
		SignCaptionGap:    1.4 * scale,
		SignCaptionHeight: 3.2 * scale,
		SignGutter:        6.8 * scale,

		ColNo:    colNo,
		ColName:  colName,
		ColQty:   colQty,
		ColPrice: colPrice,
		ColSum:   colSum,

		RuleBold: 0.44,
		RuleThin: 0.26,
		RuleHair: 0.10,
	}
}

// EN: Function `layoutTypographicAct`.
//
// EN: What it does: layoutTypographicAct walks the whole blank once and either measures it or draws it.
//
// EN: Key points: measuring and drawing share this single walk, so the fitted scale always matches what ends up on
// EN: the page; the function returns the y coordinate right below the last element.
func layoutTypographicAct(pdf *gofpdf.Fpdf, data actPDFData, layout typoActLayout, render bool) float64 {
	left := layout.Left
	width := layout.ContentWidth
	y := layout.Top

	// Title block: letter-spaced "АКТ" above the subtitle and a full width rule.
	if render {
		pdf.SetFont("Mercel", "B", layout.TitleFont)
		drawTrackedText(pdf, left, y, layout.TitleHeight, "АКТ", layout.TitleTracking)
	}
	y += layout.TitleHeight + layout.TitleGap

	if render {
		pdf.SetFont("Mercel", "", layout.SubtitleFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(width, layout.SubtitleHeight, "приёма-сдачи оказанных услуг", "", 0, "L", false, 0, "")
	}
	y += layout.SubtitleHeight + layout.SubtitleGap
	if render {
		pdf.SetLineWidth(layout.RuleBold)
		pdf.Line(left, y, left+width, y)
	}
	y += layout.RuleGap

	// Document number, city and date on one line.
	if render {
		pdf.SetFont("Mercel", "", layout.MetaFont)
		pdf.SetXY(left, y)
		pdf.CellFormat(width/2, layout.MetaHeight, fmt.Sprintf("№ %d     г. Санкт-Петербург", data.ActNumber), "", 0, "L", false, 0, "")
		pdf.CellFormat(width/2, layout.MetaHeight, russianActDate(data.GeneratedAt), "", 0, "R", false, 0, "")
	}
	y += layout.MetaHeight + layout.MetaGap

	// Parties.
	y = drawTypoSection(pdf, layout, y, "СТОРОНЫ", render)
	parties := typoPartyParagraphs(data)
	for index, paragraph := range parties {
		gap := layout.BodyGap
		if index == len(parties)-1 {
			gap = layout.BasisGap
		}
		y = drawActParagraph(pdf, layout.Left, layout.ContentWidth, y, paragraph, "", layout.BodyFont, layout.BodyLineHeight, "J", render) + gap
	}

	// Price list.
	y = drawTypoSection(pdf, layout, y, "ПЕРЕЧЕНЬ УСЛУГ", render)
	y = drawTypoActTable(pdf, layout, data, y, render) + layout.TableGap

	// Total in words.
	y = drawActParagraph(pdf, layout.Left, layout.ContentWidth, y, actTotalText(data), "", layout.TotalsFont, layout.TotalsLineHeight, "J", render) + layout.TotalsGap

	// Closing statement.
	y = drawActParagraph(pdf, layout.Left, layout.ContentWidth, y, typoClosingText(), "", layout.ClosingFont, layout.ClosingLineHeight, "J", render) + layout.ClosingGap

	// Signatures.
	y = drawTypoSignatures(pdf, layout, data, y, render)
	return y
}

// EN: Function `drawTypoSection`.
//
// EN: What it does: drawTypoSection prints one of the small letter-spaced section labels of blank №3.
//
// EN: Key points: returns the y coordinate of the next element, including the spacing below the label.
func drawTypoSection(pdf *gofpdf.Fpdf, layout typoActLayout, y float64, title string, render bool) float64 {
	if render {
		pdf.SetFont("Mercel", "B", layout.SectionFont)
		drawTrackedText(pdf, layout.Left, y, layout.SectionHeight, title, layout.SectionTracking)
	}
	return y + layout.SectionHeight + layout.SectionGap
}

// EN: Function `drawTypoActTable`.
//
// EN: What it does: drawTypoActTable renders the five column price list with rule-only borders.
//
// EN: Key points: the table has no vertical lines — a heavy rule under the header, hairlines between rows and a
// EN: heavy rule above the total, exactly like the source blank.
func drawTypoActTable(pdf *gofpdf.Fpdf, layout typoActLayout, data actPDFData, y float64, render bool) float64 {
	left := layout.Left
	padX := layout.TableCellPadX

	xNo := left
	xName := xNo + layout.ColNo
	xQty := xName + layout.ColName
	xPrice := xQty + layout.ColQty
	xSum := xPrice + layout.ColPrice
	right := left + layout.ContentWidth

	if render {
		pdf.SetFont("Mercel", "B", layout.TableHeaderFont)
		headerY := y + ((layout.TableHeaderHeight - layout.SectionHeight) / 2)
		drawTrackedText(pdf, xNo+padX, headerY, layout.SectionHeight, "№", layout.TableHeaderTracking)
		drawTrackedText(pdf, xName+padX, headerY, layout.SectionHeight, "НАИМЕНОВАНИЕ УСЛУГИ", layout.TableHeaderTracking)
		drawTypoTrackedRight(pdf, xQty+layout.ColQty-padX, headerY, layout.SectionHeight, "КОЛ-ВО", layout.TableHeaderTracking)
		drawTypoTrackedRight(pdf, xPrice+layout.ColPrice-padX, headerY, layout.SectionHeight, "ЦЕНА", layout.TableHeaderTracking)
		drawTypoTrackedRight(pdf, xSum+layout.ColSum-padX, headerY, layout.SectionHeight, "СУММА, РУБ.", layout.TableHeaderTracking)
	}
	y += layout.TableHeaderHeight
	if render {
		pdf.SetLineWidth(layout.RuleBold)
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
			pdf.SetXY(xNo+padX, y)
			pdf.CellFormat(layout.ColNo-padX, rowHeight, fmt.Sprintf("%d", index+1), "", 0, "L", false, 0, "")

			textY := y + ((rowHeight - (float64(len(lines)) * layout.TableLineHeight)) / 2)
			for lineIndex, line := range lines {
				pdf.SetXY(xName+padX, textY+(float64(lineIndex)*layout.TableLineHeight))
				pdf.CellFormat(layout.ColName-(padX*2), layout.TableLineHeight, line, "", 0, "L", false, 0, "")
			}

			pdf.SetXY(xQty, y)
			pdf.CellFormat(layout.ColQty-padX, rowHeight, formatDocumentQuantity(item.Quantity, item.Unit), "", 0, "C", false, 0, "")
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
		pdf.SetLineWidth(layout.RuleBold)
		pdf.Line(left, y, right, y)
		pdf.SetFont("Mercel", "B", layout.TableTotalFont)
		labelY := y + ((layout.TableTotalHeight - layout.SectionHeight) / 2)
		drawTypoTrackedRight(pdf, xPrice+layout.ColPrice-padX, labelY, layout.SectionHeight, "ИТОГО", layout.TableHeaderTracking)
		pdf.SetFont("Mercel", "B", layout.TableTotalValueFont)
		pdf.SetXY(xSum, y)
		pdf.CellFormat(layout.ColSum-padX, layout.TableTotalHeight, formatThousands(data.TotalAmount), "", 0, "R", false, 0, "")
	}
	return y + layout.TableTotalHeight
}

// EN: Function `drawTypoSignatures`.
//
// EN: What it does: drawTypoSignatures renders the two signature columns closing blank №3.
//
// EN: Key points: both columns are the same height, so the returned bottom edge is what the fit check compares
// EN: against the printable area.
func drawTypoSignatures(pdf *gofpdf.Fpdf, layout typoActLayout, data actPDFData, y float64, render bool) float64 {
	columnWidth := (layout.ContentWidth - layout.SignGutter) / 2
	leftColumn := layout.Left
	rightColumn := layout.Left + columnWidth + layout.SignGutter

	if render {
		pdf.SetLineWidth(layout.RuleThin)
		pdf.Line(leftColumn, y, leftColumn+columnWidth, y)
		pdf.Line(rightColumn, y, rightColumn+columnWidth, y)
	}

	bottom := drawTypoSignatureColumn(pdf, layout, leftColumn, y, "ИСПОЛНИТЕЛЬ", shortEmployeeSignatureName(data.EmployeeFullName), render)
	drawTypoSignatureColumn(pdf, layout, rightColumn, y, "ЗАКАЗЧИК", resolveContractDirectorShort(data.ContractTitle), render)
	return bottom
}

// EN: Function `drawTypoSignatureColumn`.
//
// EN: What it does: drawTypoSignatureColumn stacks the role label, the short name, the signature rule and its caption.
//
// EN: Key points: the short name is printed above the rule so the exported act is filled in automatically.
func drawTypoSignatureColumn(pdf *gofpdf.Fpdf, layout typoActLayout, x float64, y float64, label string, name string, render bool) float64 {
	cursor := y + layout.SignPadTop
	if render {
		pdf.SetFont("Mercel", "B", layout.SignLabelFont)
		drawTrackedText(pdf, x, cursor, layout.SignLabelHeight, label, layout.SignLabelTracking)
	}
	cursor += layout.SignLabelHeight + layout.SignNameGap

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

// EN: Function `drawTrackedText`.
//
// EN: What it does: drawTrackedText prints text rune by rune with extra letter spacing.
//
// EN: Key points: gofpdf has no character spacing option, and blank №3 relies on tracked capitals for its title,
// EN: section labels and table headings.
func drawTrackedText(pdf *gofpdf.Fpdf, x float64, y float64, height float64, text string, tracking float64) float64 {
	cursor := x
	for _, symbol := range text {
		glyph := string(symbol)
		glyphWidth := pdf.GetStringWidth(glyph)
		pdf.SetXY(cursor, y)
		pdf.CellFormat(glyphWidth, height, glyph, "", 0, "L", false, 0, "")
		cursor += glyphWidth + tracking
	}
	return trackedTextWidth(pdf, text, tracking)
}

// EN: Function `drawTypoTrackedRight`.
//
// EN: What it does: drawTypoTrackedRight prints tracked text ending exactly at the given right edge.
//
// EN: Key points: the numeric columns of blank №3 are right aligned, headers included.
func drawTypoTrackedRight(pdf *gofpdf.Fpdf, right float64, y float64, height float64, text string, tracking float64) {
	drawTrackedText(pdf, right-trackedTextWidth(pdf, text, tracking), y, height, text, tracking)
}

// EN: Function `trackedTextWidth`.
//
// EN: What it does: trackedTextWidth measures how wide tracked text will be with the current font.
//
// EN: Key points: the trailing gap after the last glyph is not part of the visible width.
func trackedTextWidth(pdf *gofpdf.Fpdf, text string, tracking float64) float64 {
	runes := []rune(text)
	if len(runes) == 0 {
		return 0
	}
	return pdf.GetStringWidth(text) + (float64(len(runes)-1) * tracking)
}

// EN: Function `typoPartyParagraphs`.
//
// EN: What it does: typoPartyParagraphs builds the "Стороны" block of blank №3 from the calculation data.
//
// EN: Key points: keeps the wording of the source blank; the customer director acts on the basis of the company
// EN: charter, which is the standard basis for a general director of an ООО.
func typoPartyParagraphs(data actPDFData) []string {
	return []string{
		fmt.Sprintf("Заказчик: ООО «%s», в лице генерального директора %s, действующего на основании Устава.", data.CustomerName, resolveContractDirectorName(data.ContractTitle)),
		fmt.Sprintf("Исполнитель: индивидуальный предприниматель %s.", strings.TrimSpace(data.EmployeeFullName)),
		fmt.Sprintf("Основание: Договор %s № %s от %s. Исполнитель выполнил, а Заказчик принял следующие услуги:", data.ContractTitle, data.ContractNumber, formatContractDateLong(data.ContractDate)),
	}
}

// EN: Function `typoClosingText`.
//
// EN: What it does: typoClosingText returns the closing statement wording used by blank №3.
//
// EN: Key points: shorter than the classic wording, matching the source document.
func typoClosingText() string {
	return "Услуги выполнены полностью и в срок. Заказчик претензий по объёму, качеству и срокам оказания услуг не имеет."
}
