import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("x-template-switch")
export class XTemplateSwitch extends LitElement {
  @property() private on: "class" | "attribute" = "class";
  @property() private target: string = "";

  @property({ attribute: "selector" }) private selector?: string;
  @property({ attribute: "closest-selector" }) private closestSelector?: string;

  constructor() {
    super();
  }

  private cases = new Map<string, HTMLTemplateElement>();
  private defaultCase: HTMLTemplateElement | null = null;
  private renderedNodes: Node[] = [];

  firstUpdated(_changedProperties: any) {
    super.firstUpdated(_changedProperties);

    let target: Element | null;
    if (this.selector) {
      target = this.querySelector(this.selector);
    } else if (this.closestSelector) {
      target = this.closest(this.closestSelector);
    } else {
      throw new Error(
        "x-template-switch: selector or closest-selector attribute required"
      );
    }
    if (target === null) {
      throw new Error();
    }
    console.log(target);

    const slot = this.shadowRoot?.querySelector<HTMLSlotElement>(`slot`);
    slot?.addEventListener("slotchange", () => {
      slot?.assignedElements().forEach((el) => {
        if (el instanceof HTMLTemplateElement) {
          let key = el.getAttribute("data-switch-case");
          if (key) {
            this.cases.set(key, el);
          } else if (el.hasAttribute("data-switch-default"))
            this.defaultCase = el;
        }
      });

      if (this.on === "class") {
        const observer = new MutationObserver(() =>
          this.updateDisplay(target!)
        );
        observer.observe(target!, {
          attributes: true,
          attributeFilter: ["class"],
        });
      }

      this.updateDisplay(target!);
    });
  }

  updateDisplay(target: Element) {
    const matchingCase = Array.from(this.cases.entries()).find(([key]) =>
      target.classList.contains(key)
    );

    // // Remove previously rendered content
    if (this.renderedNodes) {
      this.renderedNodes.forEach((n) => this.shadowRoot?.removeChild(n));
      this.renderedNodes = [];
    }

    // Render the matched case or default case
    const templateToRender = matchingCase ? matchingCase[1] : this.defaultCase;
    if (templateToRender) {
      const fragment = templateToRender.content.cloneNode(true);
      this.renderedNodes = Array.from(fragment.childNodes);
      this.renderedNodes.forEach((n) => this.shadowRoot?.appendChild(n));
    }
  }

  render() {
    return html`<slot></slot>`;
  }
}
