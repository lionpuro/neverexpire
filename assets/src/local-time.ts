class LocalTime extends HTMLElement {
	public datetime?: string;
	public dateonly: boolean = false;

	public constructor() {
		super();
	}

	public static get observedAttributes(): string[] {
		return ["datetime", "dateonly"];
	}

	public attributeChangedCallback(
		name: string,
		_oldValue: string,
		newValue: string,
	) {
		switch (name) {
			case "datetime":
				this.datetime = newValue;
				break;
			case "dateonly":
				this.dateonly = newValue === "true" ? true : false;
				break;
		}
	}

	connectedCallback() {
		if (!this.datetime) {
			console.warn("missing attribute datetime");
			return;
		}
		const date = new Date(this.datetime);
		if (isNaN(date.getTime())) {
			if (this.datetime !== "n/a") {
				console.error("local-time: invalid datetime input");
			}
			return;
		}
		const datestr = localeString(date, this.dateonly);
		if (datestr === "Invalid Date") {
			console.error("local-time: invalid datetime input");
			return;
		}
		this.textContent = datestr;
	}
}

function localeString(date: Date, dateonly: boolean): string {
	if (dateonly) {
		return date.toLocaleDateString(window.navigator.language);
	}
	return date.toLocaleString(window.navigator.language);
}

if (!customElements.get("local-time")) {
	customElements.define("local-time", LocalTime);
}
