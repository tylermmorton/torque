//go:build browser

package pageobjects

import "github.com/mxschmitt/playwright-go"

type Element struct {
	locator playwright.Locator
}

func NewElement(locator playwright.Locator) *Element {
	return &Element{locator: locator}
}

func (e *Element) Locator() playwright.Locator { return e.locator }

func (e *Element) TextContent() (string, error) {
	return e.locator.TextContent()
}

func (e *Element) WaitFor() error {
	return e.locator.WaitFor()
}

type ButtonElement struct{ Element }

func NewButtonElement(locator playwright.Locator) *ButtonElement {
	return &ButtonElement{Element{locator: locator}}
}

func (e *ButtonElement) Click() error {
	return e.locator.Click()
}

func (e *ButtonElement) IsDisabled() (bool, error) {
	return e.locator.IsDisabled()
}

type InputElement struct{ Element }

func NewInputElement(locator playwright.Locator) *InputElement {
	return &InputElement{Element{locator: locator}}
}

func (e *InputElement) Fill(value string) error {
	return e.locator.Fill(value)
}

func (e *InputElement) Value() (string, error) {
	return e.locator.InputValue()
}

type AnchorElement struct{ Element }

func NewAnchorElement(locator playwright.Locator) *AnchorElement {
	return &AnchorElement{Element{locator: locator}}
}

func (e *AnchorElement) Href() (string, error) {
	return e.locator.GetAttribute("href")
}
