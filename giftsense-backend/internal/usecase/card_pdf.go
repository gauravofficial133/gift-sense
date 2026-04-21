package usecase

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/giftsense/backend/assets/fonts"
	"github.com/giftsense/backend/internal/domain"
	"github.com/signintech/gopdf"
)

func RenderPDF(palette domain.CardPalette, content domain.CardContent) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 297.638, H: 419.528}})

	if err := pdf.AddTTFFontData("playfair-bold", fonts.PlayfairBold); err != nil {
		return nil, fmt.Errorf("%w: load playfair font: %s", domain.ErrCardRenderFailed, err.Error())
	}
	if err := pdf.AddTTFFontData("inter-regular", fonts.InterRegular); err != nil {
		return nil, fmt.Errorf("%w: load inter font: %s", domain.ErrCardRenderFailed, err.Error())
	}

	pdf.AddPage()

	bgR, bgG, bgB, err := hexToRGB(palette.Background)
	if err != nil {
		return nil, fmt.Errorf("%w: background color: %s", domain.ErrCardRenderFailed, err.Error())
	}
	pdf.SetFillColor(bgR, bgG, bgB)
	pdf.RectFromUpperLeftWithStyle(0, 0, 297.638, 419.528, "F")

	borderR, borderG, borderB, err := hexToRGB(palette.Primary)
	if err != nil {
		return nil, fmt.Errorf("%w: border color: %s", domain.ErrCardRenderFailed, err.Error())
	}
	pdf.SetStrokeColor(borderR, borderG, borderB)
	pdf.SetLineWidth(2)
	pdf.RectFromUpperLeftWithStyle(8, 8, 281.638, 403.528, "D")

	if err := pdf.SetFont("playfair-bold", "", 16); err != nil {
		return nil, fmt.Errorf("%w: set headline font: %s", domain.ErrCardRenderFailed, err.Error())
	}
	primR, primG, primB, _ := hexToRGB(palette.Primary)
	pdf.SetTextColor(primR, primG, primB)
	pdf.SetXY(20, 90)
	if err := pdf.Cell(nil, content.Headline); err != nil {
		return nil, fmt.Errorf("%w: headline: %s", domain.ErrCardRenderFailed, err.Error())
	}

	if err := pdf.SetFont("inter-regular", "", 10); err != nil {
		return nil, fmt.Errorf("%w: set body font: %s", domain.ErrCardRenderFailed, err.Error())
	}
	inkR, inkG, inkB, _ := hexToRGB(palette.Ink)
	pdf.SetTextColor(inkR, inkG, inkB)
	bodyLines := wrapTextPDF(content.Body, 52)
	y := 115.0
	for _, line := range bodyLines {
		pdf.SetXY(20, y)
		if err := pdf.Cell(nil, line); err != nil {
			return nil, fmt.Errorf("%w: body line: %s", domain.ErrCardRenderFailed, err.Error())
		}
		y += 14
	}

	mutR, mutG, mutB, _ := hexToRGB(palette.Muted)
	pdf.SetTextColor(mutR, mutG, mutB)
	if err := pdf.SetFont("inter-regular", "", 9); err != nil {
		return nil, fmt.Errorf("%w: set closing font: %s", domain.ErrCardRenderFailed, err.Error())
	}
	pdf.SetXY(20, y+10)
	if err := pdf.Cell(nil, content.Closing); err != nil {
		return nil, fmt.Errorf("%w: closing: %s", domain.ErrCardRenderFailed, err.Error())
	}
	pdf.SetXY(20, y+24)
	if err := pdf.Cell(nil, content.Signature); err != nil {
		return nil, fmt.Errorf("%w: signature: %s", domain.ErrCardRenderFailed, err.Error())
	}

	accR, accG, accB, _ := hexToRGB(palette.Accent)
	pdf.SetTextColor(accR, accG, accB)
	if err := pdf.SetFont("playfair-bold", "", 11); err != nil {
		return nil, fmt.Errorf("%w: set recipient font: %s", domain.ErrCardRenderFailed, err.Error())
	}
	pdf.SetXY(20, 385)
	if err := pdf.Cell(nil, "For "+content.Recipient); err != nil {
		return nil, fmt.Errorf("%w: recipient: %s", domain.ErrCardRenderFailed, err.Error())
	}

	var buf bytes.Buffer
	if _, err := pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("%w: write pdf: %s", domain.ErrCardRenderFailed, err.Error())
	}
	return buf.Bytes(), nil
}

func wrapTextPDF(text string, maxChars int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]

	for _, word := range words[1:] {
		if len(current)+1+len(word) <= maxChars {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	lines = append(lines, current)
	return lines
}

func hexToRGB(hex string) (r, g, b uint8, err error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: #%s", hex)
	}
	rv, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid red component: %w", err)
	}
	gv, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid green component: %w", err)
	}
	bv, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid blue component: %w", err)
	}
	return uint8(rv), uint8(gv), uint8(bv), nil
}
