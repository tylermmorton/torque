import { LitElement, html, PropertyValues, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";

@customElement("x-tabbed-view")
export class XTabbedView extends LitElement {
  static styles = css`
    /* Customize the editor styling */

    .container {
      border: 1px solid rgb(76 76 107 / 0.2);
      border-top-left-radius: 0.375rem;
      border-top-right-radius: 0.375rem;
      font-family: Fira Code, monospace;
    }

    .header {
      height: 40px;
      background-color: #13121c;
      border-top-left-radius: 0.375rem;
      border-top-right-radius: 0.375rem;
      display: flex;
      flex-direction: row;
      justify-content: space-between;
    }

    .editor {
      overflow-y: scroll;
      height: fit-content; /* Set a height for the editor */
      max-height: 700px;
      border-top: 1px solid rgb(76 76 107 / 0.2);
    }

    .tab-list {
      display: flex;
      flex-direction: row;
      background-color: transparent;
      border: none;
      height: 100%;
      width: 100%;
    }

    .tab {
      display: block;
    }

    .tab.hidden {
      display: none;
    }

    .tab-button {
      background-color: transparent;
      border: none;
      padding-right: 16px;
      padding-left: 16px;
      color: white;
      height: 100%;
      font-weight: bold;
      font-family: Fira Code, monospace;
      font-size: 12px;
    }

    .tab-button:first-child {
      border-top-left-radius: 0.375rem;
    }

    .tab-button:hover {
      background-color: #3b3b54;
      cursor: pointer;
    }

    .tab-button.active {
      background-color: #3b3b54;
      border-bottom: 1px solid rgb(190 144 245);
    }

    ::-webkit-scrollbar {
      width: 10px;
      height: 8px;
    }

    ::-webkit-scrollbar-track {
      background: #13121c;
    }

    ::-webkit-scrollbar-thumb {
      background: #3b3b54;
      border: 1px transparent;
    }

    ::-webkit-scrollbar-thumb:hover {
      background: #44445f;
      cursor: grab;
    }

    ::-webkit-scrollbar-thumb:hover:active {
      background: #44445f;
      cursor: grabbing;
    }
  `;

  @state()
  private tabs?: Element[];

  @state()
  private selected: number = 0;

  constructor() {
    super();
  }

  firstUpdated(_changedProperties: any) {
    super.firstUpdated(_changedProperties);
    this.tabs = this.shadowRoot?.querySelector(`slot`)?.assignedElements();
    this.updateSelected(0);
  }

  updateSelected(index: number) {
    this.selected = index;

    this.tabs?.forEach((tab, i) => {
      if (i === this.selected) {
        tab.classList.remove("hidden");
      } else {
        tab.classList.add("hidden");
      }
    });
  }

  render() {
    return html`
      <div class="container">
        <div class="header">
          <div id="tab-list">
            ${this.tabs?.map((tab, i) => {
              return html`<button
                class="tab-button ${this.selected === i ? "active" : ""}"
                @click="${() => this.updateSelected(i)}"
              >
                ${tab.getAttribute("name")}
              </button>`;
            })}
          </div>
        </div>
        <slot></slot>
      </div>
    `;
  }
}
