import { LitElement, html } from "lit";
import { customElement, property, state } from "lit/decorators.js";

@customElement("x-stargazers")
export class XStargazers extends LitElement {
  @property({ type: String })
  repo = "";

  @state()
  private stars: number = 0;

  constructor() {
    super();
  }

  connectedCallback() {
    super.connectedCallback();
    this.fetchStars();
  }

  async fetchStars() {
    if (!this.repo) return;

    try {
      const response = await fetch(`https://api.github.com/repos/${this.repo}`);
      const data = await response.json();
      this.stars = data.stargazers_count ?? 0;
    } catch (error) {
      console.error("Error fetching star count:", error);
    }
  }

  render() {
    return html`${this.stars}`;
  }
}
