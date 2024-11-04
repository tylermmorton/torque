import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";
import { EditorView, basicSetup } from "codemirror";
import { EditorState, Compartment } from "@codemirror/state";
import { go as langGo } from "@codemirror/lang-go";
import { html as langHtml } from "@codemirror/lang-html";
import { tags } from "@lezer/highlight";
import { HighlightStyle } from "@codemirror/language";
import { syntaxHighlighting } from "@codemirror/language";

const raisinHighlightStyle = HighlightStyle.define([
  { tag: tags.keyword, color: "#fc6" },
  { tag: tags.literal, color: "#99C1B9" },
  { tag: tags.typeName, color: "#81D7C6" },
  { tag: tags.comment, color: "#595959", fontStyle: "italic" },
  { tag: tags.operator, color: "#EEC0C6" },
]);

@customElement("x-code-editor")
export class XCodeEditor extends LitElement {
  static styles = css`
    /* Customize the editor styling */

    .container {
      border: 1px solid rgb(76 76 107 / 0.2);
      border-top-left-radius: 0.375rem;
      border-top-right-radius: 0.375rem;
      font-family: Fira Code, monospace;
    }

    .footer {
      height: 15px;
      background-color: #13121c;
      display: flex;
      flex-direction: row;
      justify-content: end;
      padding: 2px 8px 2px 8px;
    }

    .footer > button {
      display: flex;
      flex-direction: row;
      align-items: center;
      gap: 4px;
      background-color: transparent;
      border: none;
      padding-right: 8px;
      padding-left: 8px;
      color: white;
      height: 15px;
      font-family: Fira Code, monospace;
      font-size: 10px;
      font-weight: lighter;
    }

    .footer > button:hover {
      background-color: #3b3b54;
      cursor: pointer;
    }

    .editor {
      overflow-y: scroll;
      height: fit-content; /* Set a height for the editor */
      max-height: 700px;
      border-top: 1px solid rgb(76 76 107 / 0.2);
    }

    .cm-gutter {
      background-color: #13121c;
    }

    .cm-line.cm-activeLine {
      background-color: #211f32;
    }

    .cm-gutterElement.cm-activeLineGutter {
      background-color: #3b3b54;
    }

    .cm-selectionBackground,
    ::selection {
      background-color: #413f64 !important;
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

  private languageCompartment = new Compartment();

  @property()
  private name?: string;
  @property()
  private code?: string;
  @property()
  private language?: string;
  @property()
  private base64?: boolean;

  private editor: EditorView | null;

  constructor() {
    super();
    this.editor = null;
  }

  firstUpdated() {
    let sourceDoc: string = "";
    if (this.code && this.base64) {
      sourceDoc = atob(this.code);
    } else if (this.code) {
      sourceDoc = this.code;
    }

    let languageSupport;
    switch (this.language) {
      case "go":
        languageSupport = langGo();
        break;
      default:
        languageSupport = langHtml();
        break;
    }

    let state = EditorState.create({
      doc: sourceDoc,
      extensions: [
        basicSetup,
        this.languageCompartment.of(languageSupport),
        syntaxHighlighting(raisinHighlightStyle),
      ],
    });

    this.editor = new EditorView({
      state,
      parent: this.renderRoot.querySelector(".editor") ?? undefined,
    });
  }

  render() {
    return html`
      <div class="container">
        <div class="editor"></div>
        <div class="footer">
          <button
            @click="${() =>
              navigator.clipboard.writeText(
                this.editor?.state.doc.toString() || ""
              )}"
          >
            copy
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="12"
              height="12"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
              <path
                d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"
              ></path>
            </svg>
          </button>
        </div>
      </div>
    `;
  }
}
