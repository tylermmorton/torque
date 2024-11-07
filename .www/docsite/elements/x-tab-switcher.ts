import { LitElement, html } from "lit";
import { customElement, state } from "lit/decorators.js";

@customElement("x-tab-switcher")
export class XTabSwitcher extends LitElement {
  @state()
  private tabs?: HTMLElement[];
  private triggers?: HTMLElement[];

  constructor() {
    super();
  }

  firstUpdated(_changedProperties: any) {
    super.firstUpdated(_changedProperties);

    const slot = this.shadowRoot?.querySelector<HTMLSlotElement>(`slot`);
    slot?.addEventListener("slotchange", () => {
      console.log("slotchange");
      const elements = slot.assignedElements({
        flatten: true,
      }) as HTMLElement[];

      // Helper function to find all elements with a specific attribute recursively
      const findElementsWithAttribute = (
        elements: HTMLElement[],
        attribute: string
      ): HTMLElement[] => {
        const result: HTMLElement[] = [];
        elements.forEach((element) => {
          if (element.hasAttribute(attribute)) {
            result.push(element);
          }
          // Recursively search children
          result.push(
            ...findElementsWithAttribute(
              Array.from(element.children) as HTMLElement[],
              attribute
            )
          );
        });
        return result;
      };

      this.tabs = findElementsWithAttribute(elements, "data-tab-name");
      this.triggers = findElementsWithAttribute(elements, "data-tab-target");

      if (this.tabs === undefined || this.tabs.length == 0) {
        throw new Error(
          "x-tab-switcher cannot find any children with selector `[data-tab-name]`"
        );
      } else if (this.triggers === undefined || this.tabs.length == 0) {
        throw new Error(
          "x-tab-switcher cannot find any children with selector `[data-tab-target]`"
        );
      }

      const triggerToTabMap = new Map<HTMLElement, HTMLElement>();
      this.triggers?.forEach((triggerEl) => {
        const target = triggerEl.getAttribute("data-tab-target");

        let found = false;
        this.tabs?.forEach((tabEl) => {
          if (tabEl.getAttribute("data-tab-name") === target) {
            triggerToTabMap.set(triggerEl, tabEl);
            found = true;
          }
        });
        if (!found) {
          throw new Error("tab to trigger mismatch");
        }
      });

      triggerToTabMap.forEach((tab, trigger) => {
        trigger.addEventListener("click", () => {
          this.updateSelected(tab, trigger);
        });
      });

      let hasDefaultActive = false;
      triggerToTabMap.forEach((tab, trigger) => {
        if (trigger.classList.contains("active")) {
          this.updateSelected(tab, trigger);
          hasDefaultActive = true;
        }
      });
      if (!hasDefaultActive) {
        const firstEntry = triggerToTabMap.entries().next().value;
        const [trigger, tab] = firstEntry;
        this.updateSelected(tab, trigger);
      }
    });
  }

  updateSelected(tab: HTMLElement, trigger: HTMLElement) {
    this.tabs?.forEach((t) => t.classList.add("hidden"));
    tab.classList.remove("hidden");

    this.triggers?.forEach((t) => t.classList.remove("active"));
    trigger.classList.add("active");
  }

  render() {
    return html`<slot></slot>`;
  }
}
