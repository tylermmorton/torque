import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";

/** Usage:
 * <x-collapsible>
 *     <div slot="trigger"> // slot prevents element from collapsing
 *         <button data-collapsible-trigger></button>
 *     </div>
 *     ... content
 *  </x-collapsible>
 *
 *  data-collapsible-trigger is an optional attribute to declare the
 *  element(s) that can open/close. If not provided the entire trigger
 *  slot is used to mount the click event listener
 *
 *  Collapsing will apply the 'hidden' class to all nested elements, except
 *  for those within the trigger slot. Change the hidden class applied with
 *  the attribute collapsed-class
 */
@customElement("x-collapsible")
export class XCollapsible extends LitElement {
  constructor() {
    super();
  }

  @property({ attribute: "apply-to-content" })
  public applyToContent: string = "hidden";

  @property({ attribute: "apply-to-trigger" })
  public applyToTrigger: string = "";

  firstUpdated(_changedProperties: any) {
    super.firstUpdated(_changedProperties);

    const trigger =
      this.shadowRoot?.querySelector<HTMLSlotElement>(`slot#trigger`);

    const content =
      this.shadowRoot?.querySelector<HTMLSlotElement>(`slot#content`);

    Promise.all([
      new Promise<void>((resolve) => {
        trigger?.addEventListener("slotchange", () => resolve(), {
          once: true,
        });
      }),
      new Promise<void>((resolve) => {
        content?.addEventListener("slotchange", () => resolve(), {
          once: true,
        });
      }),
    ]).then(() => {
      let found = false;
      trigger?.assignedElements().forEach((el) => {
        el.querySelectorAll<HTMLElement>("[data-collapsible-trigger]").forEach(
          (el) => {
            el.addEventListener("click", () => {
              content
                ?.assignedElements()
                .forEach((el) => el.classList.toggle(this.applyToContent));
            });
            el.addEventListener("click", () =>
              el.classList.toggle(this.applyToTrigger)
            );
            found = true;
          }
        );
      });

      if (!found) {
        trigger?.addEventListener("click", () => {
          content
            ?.assignedElements()
            .forEach((el) => el.classList.toggle(this.applyToContent));
        });
        trigger?.addEventListener("click", () =>
          trigger?.classList.toggle(this.applyToTrigger)
        );
      }
    });
  }

  render() {
    return html` <slot id="trigger" name="trigger"></slot>
      <slot id="content"></slot>`;
  }
}
