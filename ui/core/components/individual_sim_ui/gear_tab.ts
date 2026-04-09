import tippy from 'tippy.js';
import i18n from '../../../i18n/config';
import { IndividualSimUI } from '../../individual_sim_ui';
import { Player } from '../../player';
import { PresetBuild, PresetGear } from '../../preset_utils';
import { EquipmentSpec, UnitStats } from '../../proto/common';
import { SavedGearSet } from '../../proto/ui';
import { Stats } from '../../proto_utils/stats';
import { EventID, TypedEvent } from '../../typed_event';
import GearPicker from '../gear_picker/gear_picker';
import { SavedDataManager } from '../saved_data_manager';
import { SimTab } from '../sim_tab';
import { GemSummary } from './gem_summary';
import { PresetConfigurationCategory, PresetConfigurationPicker } from './preset_configuration_picker';
import { PresetGroupPicker, PresetGroupItem } from './preset_group_picker';

export class GearTab extends SimTab {
	protected simUI: IndividualSimUI<any>;

	readonly leftPanel: HTMLElement;
	readonly rightPanel: HTMLElement;

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<any>) {
		super(parentElem, simUI, { identifier: 'gear-tab', title: i18n.t('gear_tab.title') });
		this.simUI = simUI;

		this.leftPanel = document.createElement('div');
		this.leftPanel.classList.add('gear-tab-left', 'tab-panel-left');

		this.rightPanel = document.createElement('div');
		this.rightPanel.classList.add('gear-tab-right', 'tab-panel-right');

		this.contentContainer.appendChild(this.leftPanel);
		this.contentContainer.appendChild(this.rightPanel);

		this.buildTabContent();
	}

	protected buildTabContent() {
		this.buildGearPickers();
		this.buildSummaryTablesContainer();

		if (this.hasGroupedPresets()) {
			this.buildGroupedPresets();
		} else {
			this.buildPresetConfigurationPicker();
			this.buildSavedGearsetPicker();
		}
	}

	private hasGroupedPresets(): boolean {
		const builds = this.simUI.individualConfig.presets.builds ?? [];
		const gear = this.simUI.individualConfig.presets.gear ?? [];
		return [...builds, ...gear].some(p => p.phase !== undefined || p.group !== undefined);
	}

	private buildGroupedPresets() {
		const groupPicker = new PresetGroupPicker(this.rightPanel, {
			storageKey: this.simUI.getPresetFilterStorageKey(),
		});

		const builds = (this.simUI.individualConfig.presets.builds ?? []).filter(build =>
			Object.keys(build).some(
				category =>
					[PresetConfigurationCategory.Gear as string].includes(category) &&
					!!build[category as keyof PresetBuild],
			),
		);

		this.simUI.sim.waitForInit().then(() => {
			// Preset Configurations section
			if (builds.length) {
				const buildItems: PresetGroupItem[] = builds.map(build => ({
					phase: build.phase,
					group: build.group,
					elem: this.makeBuildChip(build, groupPicker),
				}));
				groupPicker.addSection(i18n.t('gear_tab.preset_configurations.title'), buildItems);
			}

			// Gear Sets section (preset portion)
			const gearPresets = this.simUI.individualConfig.presets.gear;
			if (gearPresets.length) {
				const gearItems: PresetGroupItem[] = gearPresets.map(presetGear => ({
					phase: presetGear.phase,
					group: presetGear.group,
					elem: this.makeGearPresetChip(presetGear),
				}));
				const gearSection = groupPicker.addSection(i18n.t('gear_tab.gear_sets.title'), gearItems);

				// Append user-saved gear manager below grouped presets
				this.buildUserSavedGearSection(gearSection);
			}

			groupPicker.init();
		});
	}

	private makeBuildChip(build: PresetBuild, groupPicker: PresetGroupPicker): HTMLElement {
		const dataElem = document.createElement('button');
		dataElem.className = 'saved-data-set-chip badge rounded-pill';

		const nameSpan = document.createElement('span');
		nameSpan.className = 'saved-data-set-name';
		nameSpan.setAttribute('role', 'button');
		nameSpan.textContent = build.name;
		nameSpan.addEventListener('click', () => {
			PresetConfigurationPicker.applyBuild(TypedEvent.nextEventID(), build, this.simUI);
			groupPicker.setFilter(build.phase);
		});
		dataElem.appendChild(nameSpan);

		// Active state tracking
		const checkActive = () => {
			const isActive = build.gear
				? EquipmentSpec.equals(build.gear.gear, this.simUI.player.getGear().asSpec())
				: false;
			dataElem.classList[isActive ? 'add' : 'remove']('active');
		};
		checkActive();
		TypedEvent.onAny([
			this.simUI.player.changeEmitter,
			this.simUI.sim.settingsChangeEmitter,
			this.simUI.sim.raid.changeEmitter,
			this.simUI.sim.encounter.changeEmitter,
		]).on(checkActive);

		return dataElem;
	}

	private makeGearPresetChip(presetGear: PresetGear): HTMLElement {
		const dataElem = document.createElement('div');
		dataElem.className = 'saved-data-set-chip badge rounded-pill';

		const nameBtn = document.createElement('button');
		nameBtn.className = 'saved-data-set-name';
		nameBtn.textContent = presetGear.name;
		dataElem.appendChild(nameBtn);

		const gearData = SavedGearSet.create({
			gear: this.simUI.sim.db.lookupEquipmentSpec(presetGear.gear).asSpec(),
			bonusStatsStats: new Stats().toProto(),
		});

		dataElem.addEventListener('click', () => {
			const eventID = TypedEvent.nextEventID();
			TypedEvent.freezeAllAndDo(() => {
				this.simUI.player.setGear(eventID, this.simUI.sim.db.lookupEquipmentSpec(presetGear.gear));
				this.simUI.player.setBonusStats(eventID, new Stats());
			});
			presetGear.onLoad?.(this.simUI.player);
		});

		// Active + enableWhen tracking
		const checkState = () => {
			const currentGear = SavedGearSet.create({
				gear: this.simUI.player.getGear().asSpec(),
				bonusStatsStats: this.simUI.player.getBonusStats().toProto(),
			});
			dataElem.classList[SavedGearSet.equals(gearData, currentGear) ? 'add' : 'remove']('active');

			if (presetGear.enableWhen) {
				dataElem.classList[presetGear.enableWhen(this.simUI.player) ? 'remove' : 'add']('disabled');
			}
		};
		checkState();
		this.simUI.player.changeEmitter.on(checkState);

		if (presetGear.tooltip) {
			tippy(dataElem, { content: presetGear.tooltip, placement: 'bottom' });
		}

		return dataElem;
	}

	private buildUserSavedGearSection(parentSection: HTMLElement) {
		const savedGearManager = new SavedDataManager<Player<any>, SavedGearSet>(parentSection, this.simUI.player, {
			label: i18n.t('gear_tab.gear_sets.gear_set'),
			nameLabel: i18n.t('gear_tab.gear_sets.gear_set_name'),
			saveButtonText: i18n.t('gear_tab.gear_sets.save_gear_set'),
			storageKey: this.simUI.getSavedGearStorageKey(),
			getData: (player: Player<any>) => {
				return SavedGearSet.create({
					gear: player.getGear().asSpec(),
					bonusStatsStats: player.getBonusStats().toProto(),
				});
			},
			setData: (eventID: EventID, player: Player<any>, newSavedGear: SavedGearSet) => {
				TypedEvent.freezeAllAndDo(() => {
					player.setGear(eventID, this.simUI.sim.db.lookupEquipmentSpec(newSavedGear.gear || EquipmentSpec.create()));
					player.setBonusStats(eventID, Stats.fromProto(newSavedGear.bonusStatsStats || UnitStats.create()));
				});
			},
			changeEmitters: [this.simUI.player.changeEmitter],
			equals: (a: SavedGearSet, b: SavedGearSet) => SavedGearSet.equals(a, b),
			toJson: (a: SavedGearSet) => SavedGearSet.toJson(a),
			fromJson: (obj: any) => SavedGearSet.fromJson(obj),
		});
		savedGearManager.loadUserData();
	}

	private buildSummaryTablesContainer() {
		const container = document.createElement('div');
		container.classList.add('summary-tables-container');
		this.leftPanel.appendChild(container);

		new GemSummary(container, this.simUI, this.simUI.player);
	}

	private buildGearPickers() {
		new GearPicker(this.leftPanel, this.simUI, this.simUI.player);
	}

	private buildPresetConfigurationPicker() {
		new PresetConfigurationPicker(this.rightPanel, this.simUI, [PresetConfigurationCategory.Gear]);
	}

	private buildSavedGearsetPicker() {
		const savedGearManager = new SavedDataManager<Player<any>, SavedGearSet>(this.rightPanel, this.simUI.player, {
			header: { title: i18n.t('gear_tab.gear_sets.title') },
			label: i18n.t('gear_tab.gear_sets.gear_set'),
			nameLabel: i18n.t('gear_tab.gear_sets.gear_set_name'),
			saveButtonText: i18n.t('gear_tab.gear_sets.save_gear_set'),
			storageKey: this.simUI.getSavedGearStorageKey(),
			getData: (player: Player<any>) => {
				return SavedGearSet.create({
					gear: player.getGear().asSpec(),
					bonusStatsStats: player.getBonusStats().toProto(),
				});
			},
			setData: (eventID: EventID, player: Player<any>, newSavedGear: SavedGearSet) => {
				TypedEvent.freezeAllAndDo(() => {
					player.setGear(eventID, this.simUI.sim.db.lookupEquipmentSpec(newSavedGear.gear || EquipmentSpec.create()));
					player.setBonusStats(eventID, Stats.fromProto(newSavedGear.bonusStatsStats || UnitStats.create()));
				});
			},
			changeEmitters: [this.simUI.player.changeEmitter],
			equals: (a: SavedGearSet, b: SavedGearSet) => SavedGearSet.equals(a, b),
			toJson: (a: SavedGearSet) => SavedGearSet.toJson(a),
			fromJson: (obj: any) => SavedGearSet.fromJson(obj),
		});

		this.simUI.sim.waitForInit().then(() => {
			savedGearManager.loadUserData();
			this.simUI.individualConfig.presets.gear.forEach(presetGear => {
				savedGearManager.addSavedData({
					name: presetGear.name,
					tooltip: presetGear.tooltip,
					isPreset: true,
					data: SavedGearSet.create({
						// Convert to gear and back so order is always the same.
						gear: this.simUI.sim.db.lookupEquipmentSpec(presetGear.gear).asSpec(),
						bonusStatsStats: new Stats().toProto(),
					}),
					enableWhen: presetGear.enableWhen,
					onLoad: presetGear.onLoad,
				});
			});
		});
	}
}
