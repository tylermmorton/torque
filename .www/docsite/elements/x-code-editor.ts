import { LitElement, html, PropertyValues, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { EditorView, basicSetup } from "codemirror";
import { EditorState, Compartment } from "@codemirror/state";
import { go } from "@codemirror/lang-go";
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

    .header {
      height: 30px;
      background-color: #13121c;
      border-top-left-radius: 0.375rem;
      border-top-right-radius: 0.375rem;
      display: flex;
      flex-direction: row;
      justify-content: end;
      align-items: center;
      padding: 2px 8px 2px 8px;
    }

    .editor {
      overflow-y: scroll;
      height: fit-content; /* Set a height for the editor */
      max-height: 700px;
      border-top: 1px solid rgb(76 76 107 / 0.2);
    }

    .copyButton {
      margin-left: auto;
      background-color: transparent;
      color: white;
      border: none;
      padding: 3px 3px 3px 3px;
      border-radius: 0.375rem;
      display: flex;
      justify-content: center;
      align-items: center;
    }

    .copyButton:hover {
      background-color: #3b3b54;
      cursor: pointer;
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

  private language = new Compartment();

  @property()
  private code?: string;
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

    let state = EditorState.create({
      doc: sourceDoc,
      extensions: [
        basicSetup,
        this.language.of(go()),
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
        <div class="header">
          <button class="copyButton">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="feather feather-copy"
            >
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
              <path
                d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"
              ></path>
            </svg>
          </button>
        </div>
        <div class="editor"></div>
      </div>
    `;
  }
}
