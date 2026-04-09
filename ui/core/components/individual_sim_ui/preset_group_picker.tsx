import { ref } from 'tsx-vanilla';

import { CURRENT_PHASE } from '../../constants/other';
import { Component } from '../component';

export interface PresetGroupItem {
	phase?: number;
	group?: string;
	elem: HTMLElement;
}

export interface PresetGroupPickerConfig {
	storageKey: string;
}

interface FilterState {
	phase: number;
}

export class PresetGroupPicker extends Component {
	private readonly config: PresetGroupPickerConfig;
	private readonly items: PresetGroupItem[] = [];
	private readonly itemSectionMap: Map<PresetGroupItem, HTMLElement> = new Map();

	private readonly phaseTabsContainer: HTMLElement;
	private readonly sectionsContainer: HTMLElement;
	private readonly onFilterCallbacks: Array<() => void> = [];

	private filterState: FilterState;
	private phases: number[] = [];
	private groups: string[] = [];

	constructor(parent: HTMLElement, config: PresetGroupPickerConfig) {
		super(parent, 'preset-group-picker');
		this.config = config;

		const phaseTabsRef = ref<HTMLDivElement>();
		const sectionsRef = ref<HTMLDivElement>();

		this.rootElem.appendChild(
			<div className="preset-group-picker-container">
				<div ref={phaseTabsRef} className="preset-group-phase-tabs" />
				<div ref={sectionsRef} className="preset-group-sections" />
			</div>,
		);

		this.phaseTabsContainer = phaseTabsRef.value!;
		this.sectionsContainer = sectionsRef.value!;

		this.filterState = this.loadFilterState();
	}

	addSection(title: string, sectionItems: PresetGroupItem[]): HTMLElement {
		this.items.push(...sectionItems);

		const sectionRef = ref<HTMLDivElement>();
		this.sectionsContainer.appendChild(
			<div ref={sectionRef} className="preset-group-section">
				<h6 className="content-block-title">{title}</h6>
				<div className="preset-group-section-body" />
			</div>,
		);

		const sectionBody = sectionRef.value!.querySelector('.preset-group-section-body') as HTMLElement;

		sectionItems.forEach(item => {
			this.itemSectionMap.set(item, sectionBody);
		});

		return sectionRef.value!;
	}

	/**
	 * Register a callback to be invoked after each filter change.
	 */
	onFilter(callback: () => void) {
		this.onFilterCallbacks.push(callback);
	}

	/**
	 * Update the active phase, then re-render.
	 */
	setFilter(phase?: number) {
		if (phase !== undefined && this.phases.includes(phase)) {
			this.filterState.phase = phase;
		}
		this.saveFilterState();
		this.renderPhaseTabs();
		this.applyFilters();
		this.onFilterCallbacks.forEach(cb => cb());
	}

	init() {
		this.extractMetadata();
		this.validateFilterState();
		this.renderPhaseTabs();
		this.applyFilters();
	}

	private extractMetadata() {
		const phaseSet = new Set<number>();
		const groupSet = new Set<string>();

		for (const item of this.items) {
			if (item.phase !== undefined) phaseSet.add(item.phase);
			if (item.group) groupSet.add(item.group);
		}

		this.phases = [...phaseSet].sort((a, b) => a - b);
		this.groups = [...groupSet];
	}

	private validateFilterState() {
		if (!this.phases.includes(this.filterState.phase)) {
			this.filterState.phase = this.phases.includes(CURRENT_PHASE)
				? CURRENT_PHASE
				: this.phases[0] ?? CURRENT_PHASE;
		}
	}

	private renderPhaseTabs() {
		this.phaseTabsContainer.replaceChildren();

		if (this.phases.length <= 1) {
			this.phaseTabsContainer.classList.add('hide');
			return;
		}

		this.phaseTabsContainer.classList.remove('hide');
		for (const phase of this.phases) {
			const tab = (
				<button
					className={`preset-group-phase-tab${phase === this.filterState.phase ? ' active' : ''}`}
					onclick={() => this.setFilter(phase)}>
					{`Phase ${phase}`}
				</button>
			);
			this.phaseTabsContainer.appendChild(tab);
		}
	}

	private applyFilters() {
		const sectionMap = new Map<HTMLElement, Map<string, HTMLElement[]>>();

		for (const item of this.items) {
			const sectionBody = this.itemSectionMap.get(item);
			if (!sectionBody) continue;

			// Phase filter
			if (item.phase !== undefined && item.phase !== this.filterState.phase) {
				continue;
			}

			const groupName = item.group || '';
			if (!sectionMap.has(sectionBody)) {
				sectionMap.set(sectionBody, new Map());
			}
			const groupMap = sectionMap.get(sectionBody)!;
			if (!groupMap.has(groupName)) {
				groupMap.set(groupName, []);
			}
			groupMap.get(groupName)!.push(item.elem);
		}

		// Clear and rebuild each section body
		const allSectionBodies = new Set(this.itemSectionMap.values());

		const hideGroupHeadings = this.groups.length <= 1;

		for (const sectionBody of allSectionBodies) {
			sectionBody.replaceChildren();
			const groupMap = sectionMap.get(sectionBody);
			if (!groupMap || groupMap.size === 0) continue;

			// Show ungrouped items first (e.g. "Pre-Raid")
			const ungroupedElems = groupMap.get('');
			if (ungroupedElems && ungroupedElems.length > 0) {
				const chipRow = <div className="saved-data-presets" />;
				ungroupedElems.forEach(el => chipRow.appendChild(el));
				sectionBody.appendChild(chipRow);
			}

			// Then show items by group
			for (const groupName of this.groups) {
				const elems = groupMap.get(groupName);
				if (!elems || elems.length === 0) continue;

				if (!hideGroupHeadings) {
					sectionBody.appendChild(
						<div className="preset-group-label">{groupName}</div>,
					);
				}
				const chipRow = <div className="saved-data-presets" />;
				elems.forEach(el => chipRow.appendChild(el));
				sectionBody.appendChild(chipRow);
			}
		}
	}

	private loadFilterState(): FilterState {
		try {
			const stored = window.localStorage.getItem(this.config.storageKey);
			if (stored) {
				const parsed = JSON.parse(stored);
				return {
					phase: parsed.phase ?? CURRENT_PHASE,
				};
			}
		} catch {
			// Ignore corrupt localStorage
		}
		return { phase: CURRENT_PHASE };
	}

	private saveFilterState() {
		window.localStorage.setItem(this.config.storageKey, JSON.stringify(this.filterState));
	}
}
