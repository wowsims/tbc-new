import {
	AuraStats as AuraStatsProto,
	Player as PlayerProto,
	PlayerStats,
	SpellStats as SpellStatsProto,
	StatWeightsResult,
	UnitMetadata as UnitMetadataProto,
} from '@generated/proto/api';
import { APLRotation, APLRotation_Type as APLRotationType, SimpleRotation } from '@generated/proto/apl';
import {
	Class,
	ConsumableType,
	ConsumesSpec,
	Cooldowns,
	Faction,
	GemColor,
	HandType,
	HealingModel,
	IndividualBuffs,
	ItemRandomSuffix,
	ItemSlot,
	Profession,
	PseudoStat,
	Race,
	Spec,
	Stat,
	TristateEffect,
	UnitReference,
	UnitStats,
	WeaponType,
} from '@generated/proto/common';
import { SimDatabase } from '@generated/proto/db';
import {
	DungeonDifficulty,
	RaidFilterOption,
	SourceFilterOption,
	UIEnchant as Enchant,
	UIGem as Gem,
	UIItem as Item,
	UIItem_FactionRestriction,
} from '@generated/proto/ui';

import * as Mechanics from '../constants/mechanics';
import { CURRENT_API_VERSION } from '../constants/other';
import { SimSettingCategories } from '../constants/sim_settings';
import type { PresetEpWeights } from '../presets/types';
import { ActionId } from '../proto/action_id';
import { Database } from '../proto/database';
import { EquippedItem } from '../proto/equipped_item';
import { Gear, ItemSwapGear } from '../proto/gear';
import { gemMatchesSocket, isUnrestrictedGem } from '../proto/gems';
import { canEquipEnchant, canEquipItem, enchantAppliesToItem, getMetaGemEffectEP, isPVPItem } from '../proto/items';
import { migrateOldProto, ProtoConversionMap } from '../proto/proto_migration';
import { specTypeFunctions, withSpec } from '../proto/spec_functions';
import type { ClassOptions, ClassSpecs, SpecClasses, SpecOptions, SpecRotation, SpecTalents, SpecTypeFunctions } from '../proto/spec_types';
import { Stats, UnitStat } from '../proto/stats';
import {
	ADAMANTITE_SHARPENING_STONE_ID,
	ADAMANTITE_WEIGHTSTONE_ID,
	AL_CATEGORY_HARD_MODE,
	emptyUnitReference,
	getTalentTreePoints,
	newUnitReference,
	raceToFaction,
} from '../proto/utils';
import { MAX_PARTY_SIZE, Party } from '../raid/party';
import { Raid } from '../raid/raid';
import { CONJURED_CONFIG, relevantConsumableOptions } from '../settings/conjured';
import { ItemSwapSettings } from '../settings/item_swap_settings';
import { Sim } from '../sim';
import { batch } from '../state/batch';
import { deleteKeyed, patchKeyed, PLAYER_FIELDS, PlayerField, PlayerSlice, seedKeyed, zeroVersions } from '../state/sim_store';
import { playerTalentStringToProto } from '../talents/factory';
import { omitDeep, stringComparator } from '../utils/collections';
import { sum } from '../utils/math';
import { WorkerProgressCallback } from '../workers/worker_pool';
import { characterSheetDebuffStats, characterSheetExposeWeaknessAgility } from './debuff_stats';
import { PlayerClass } from './player_class';
import { PlayerSpec } from './player_spec';
import { PlayerSpecs } from './specs';

export interface AuraStats {
	data: AuraStatsProto;
	id: ActionId;
}
export interface SpellStats {
	data: SpellStatsProto;
	id: ActionId;
}

export class UnitMetadata {
	private name: string;
	private auras: Array<AuraStats>;
	private spells: Array<SpellStats>;

	constructor() {
		this.name = '';
		this.auras = [];
		this.spells = [];
	}

	getName(): string {
		return this.name;
	}

	getAuras(): Array<AuraStats> {
		return this.auras.slice();
	}

	getSpells(): Array<SpellStats> {
		return this.spells.slice();
	}

	// Returns whether any updates were made.
	async update(metadata: UnitMetadataProto): Promise<boolean> {
		let newSpells = metadata!.spells.map(spell => {
			return {
				data: spell,
				id: ActionId.fromProto(spell.id!),
			};
		});
		let newAuras = metadata!.auras.map(aura => {
			return {
				data: aura,
				id: ActionId.fromProto(aura.id!),
			};
		});

		await Promise.all([...newSpells, ...newAuras].map(newSpell => newSpell.id.fill().then(newId => (newSpell.id = newId))));

		newSpells = newSpells.sort((a, b) => stringComparator(a.id.name, b.id.name));
		newAuras = newAuras.sort((a, b) => stringComparator(a.id.name, b.id.name));

		let anyUpdates = false;
		if (metadata.name != this.name) {
			this.name = metadata.name;
			anyUpdates = true;
		}
		if (newSpells.length != this.spells.length || newSpells.some((newSpell, i) => !newSpell.id.equals(this.spells[i].id))) {
			this.spells = newSpells;
			anyUpdates = true;
		}
		if (newAuras.length != this.auras.length || newAuras.some((newAura, i) => !newAura.id.equals(this.auras[i].id))) {
			this.auras = newAuras;
			anyUpdates = true;
		}

		return anyUpdates;
	}
}

export class UnitMetadataList {
	private metadatas: Array<UnitMetadata>;

	constructor() {
		this.metadatas = [];
	}

	async update(newMetadatas: Array<UnitMetadataProto>): Promise<boolean> {
		const oldLen = this.metadatas.length;

		if (newMetadatas.length > oldLen) {
			for (let i = oldLen; i < newMetadatas.length; i++) {
				this.metadatas.push(new UnitMetadata());
			}
		} else if (newMetadatas.length < oldLen) {
			this.metadatas = this.metadatas.slice(0, newMetadatas.length);
		}

		const anyUpdates = await Promise.all(newMetadatas.map((metadata, i) => this.metadatas[i].update(metadata)));

		return oldLen != this.metadatas.length || anyUpdates.some(v => v);
	}

	asList(): Array<UnitMetadata> {
		return this.metadatas.slice();
	}
}

export interface MeleeCritCapInfo {
	meleeCrit: number;
	meleeHit: number;
	expertise: number;
	suppression: number;
	glancing: number;
	debuffCrit: number;
	hasOffhandWeapon: boolean;
	meleeHitCap: number;
	expertiseCap: number;
	remainingMeleeHitCap: number;
	remainingExpertiseCap: number;
	baseCritCap: number;
	specSpecificOffset: number;
	playerCritCapDelta: number;
}

export type AutoRotationGenerator<SpecType extends Spec> = (player: Player<SpecType>) => APLRotation;
export type SimpleRotationGenerator<SpecType extends Spec> = (
	player: Player<SpecType>,
	simpleRotation: SpecRotation<SpecType>,
	cooldowns: Cooldowns,
) => APLRotation;

export interface PlayerConfig<SpecType extends Spec> {
	autoRotation: AutoRotationGenerator<SpecType>;
	simpleRotation?: SimpleRotationGenerator<SpecType>;
	hiddenMCDs?: Array<number>; // spell IDs for any MCDs that should be omitted from the Simple Cooldowns UI
}

// The subset of the per-spec UI config (IndividualSimUIConfig) that the domain
// layer consumes. Spec configs registered via registerSpecConfig satisfy this
// structurally; the full UI config type stays in individual_sim_ui.
export interface SpecConfigData<SpecType extends Spec> extends PlayerConfig<SpecType> {
	// Override for required talent rows. If unset, all rows [0..5] are required.
	requiredTalentRows?: number[];
	epStats: Array<Stat>;
	consumableStats?: Array<Stat>;
	gemStats?: Array<Stat>;
	// Per-spec override for the default EP ratios; must be exactly numEpRatios long.
	epRatios?: Array<number>;
	displayStats?: Array<UnitStat>;
	includeBuffDebuffInputs?: Array<Stat | PseudoStat>;
	excludeBuffDebuffInputs?: Array<Stat | PseudoStat>;
	presets: {
		epWeights: Array<PresetEpWeights>;
	};
}

const SPEC_CONFIGS: Partial<Record<Spec, PlayerConfig<any>>> = {};

export function registerSpecConfig<SpecType extends Spec>(spec: SpecType, config: PlayerConfig<SpecType>) {
	SPEC_CONFIGS[spec] = config;
}

export function getSpecConfig<SpecType extends Spec>(spec: SpecType): PlayerConfig<SpecType> {
	const config = SPEC_CONFIGS[spec] as PlayerConfig<SpecType>;
	if (!config) {
		throw new Error('No config registered for Spec: ' + spec);
	}
	return config;
}

// Manages all the gear / consumes / other settings for a single Player.
export class Player<SpecType extends Spec> {
	readonly sim: Sim;
	private party: Party | null;
	private raid: Raid | null;

	readonly playerSpec: PlayerSpec<SpecType>;
	readonly playerClass: PlayerClass<SpecClasses<SpecType>>;

	// Settings fields live in the sim store (players[storeKey]); see PlayerSlice.
	//private bulkEquipmentSpec: BulkEquipmentSpec = BulkEquipmentSpec.create();
	itemSwapSettings: ItemSwapSettings;
	private aplRotation_: APLRotation = APLRotation.create();

	// Read-only access: the returned rotation is the live object — do NOT mutate
	// it directly; route writes through setAplRotation/modifyAplRotation.
	get aplRotation(): APLRotation {
		return this.aplRotation_;
	}

	private healingEnabled = false;

	private readonly autoRotationGenerator: AutoRotationGenerator<SpecType> | null = null;
	private readonly simpleRotationGenerator: SimpleRotationGenerator<SpecType> | null = null;
	readonly hiddenMCDs: Array<number>;

	private itemEPCache = new Array<Map<string, number>>();
	private gemEPCache = new Map<number, number>();
	private randomSuffixEPCache = new Map<number, number>();
	private enchantEPCache = new Map<number, number>();
	private talents: SpecTalents<SpecType> | null = null;
	private specConfig: SpecConfigData<SpecType>;

	readonly specTypeFunctions: SpecTypeFunctions<SpecType>;

	private static readonly numEpRatios = 6;
	private metadata: UnitMetadata = new UnitMetadata();
	private petMetadatas: UnitMetadataList = new UnitMetadataList();

	private static nextStoreKey = 0;
	// Key of this player's slice in sim.store; unique per Player instance.
	// Call dispose() when an instance is discarded (see state/README.md).
	readonly storeKey = Player.nextStoreKey++;
	private readonly unsubscribers: Array<() => void> = [];
	private disposed = false;

	private slice() {
		return this.sim.store.getState().players[this.storeKey];
	}

	// Writes `patch` and bumps the given version counters in one store write.
	// The bump is what subscribers watch, so a setter notifies exactly when it
	// calls write() — guard logic stays in the setters. Counter-only fields
	// (rotation, itemSwap, epRefStat) are bumped with an empty patch.
	private write(patch: Partial<Omit<PlayerSlice, 'v'>>, bumps: ReadonlyArray<PlayerField>) {
		patchKeyed(this.sim.store, 'players', this.storeKey, patch, bumps);
	}

	// Writes one field and bumps its version.
	private patch<F extends Exclude<PlayerField, 'rotation' | 'itemSwap' | 'epRefStat'>>(field: F, value: PlayerSlice[F]) {
		this.write({ [field]: value } as Partial<Omit<PlayerSlice, 'v'>>, [field]);
	}

	// Signals a rotation change made in place on `aplRotation` (APL editor).
	touchRotation() {
		this.write({}, ['rotation']);
	}

	// Item-swap fields share one version counter (ItemSwapSettings facade).
	patchItemSwap(patch: { itemSwapEnabled?: boolean; itemSwapGear?: ItemSwapGear; itemSwapBonusStats?: Stats }) {
		this.write(patch, ['itemSwap']);
	}

	getItemSwapField<F extends 'itemSwapEnabled' | 'itemSwapGear' | 'itemSwapBonusStats'>(field: F): PlayerSlice[F] {
		return this.slice()[field];
	}

	constructor(spec: PlayerSpec<SpecType>, sim: Sim) {
		this.sim = sim;
		this.party = null;
		this.raid = null;

		this.playerSpec = spec;
		this.playerClass = PlayerSpecs.getPlayerClass(spec);

		this.specTypeFunctions = specTypeFunctions[this.getSpec()] as SpecTypeFunctions<SpecType>;

		// Seed this player's slice (field initialization, not a change).
		seedKeyed(this.sim.store, 'players', this.storeKey, {
			name: '',
			race: this.playerClass.races[0],
			profession1: 0,
			profession2: 0,
			buffs: IndividualBuffs.create(),
			consumables: ConsumesSpec.create(),
			bonusStats: new Stats(),
			gear: new Gear({}),
			talentsString: '',
			specOptions: this.specTypeFunctions.optionsCreate(),
			reactionTime: 0,
			channelClipDelay: 0,
			inFrontOfTarget: false,
			distanceFromTarget: 0,
			healingModel: HealingModel.create(),
			epWeights: new Stats(),
			epRatios: new Array<number>(Player.numEpRatios).fill(0),
			currentStats: PlayerStats.create(),
			itemSwapEnabled: false,
			itemSwapGear: new ItemSwapGear({}),
			itemSwapBonusStats: new Stats(),
			dpsRefStat: undefined,
			healRefStat: undefined,
			tankRefStat: undefined,
			v: zeroVersions(PLAYER_FIELDS),
		});

		this.specConfig = getSpecConfig<SpecType>(this.getSpec()) as SpecConfigData<SpecType>;

		this.autoRotationGenerator = this.specConfig.autoRotation;
		if (this.specConfig.simpleRotation) {
			this.simpleRotationGenerator = this.specConfig.simpleRotation;
		} else {
			this.simpleRotationGenerator = null;
		}
		this.hiddenMCDs = this.specConfig.hiddenMCDs || new Array<number>();

		for (let i = 0; i < ItemSlot.ItemSlotRanged + 1; ++i) {
			this.itemEPCache[i] = new Map();
		}

		this.itemSwapSettings = new ItemSwapSettings(this);
	}

	// Releases this instance's store subscriptions and slices. Only call when
	// the Player is genuinely discarded (replaced by a different instance and
	// referenced nowhere else) — a player moved between parties is NOT discarded.
	dispose() {
		if (this.disposed) return;
		this.disposed = true;
		this.unsubscribers.splice(0).forEach(u => u());
		// Drop the slices on the next tick: pickers bound to this player are
		// replaced by the UI's own (gated) composition subscribers first, so they
		// are already off the DOM — and self-dispose — when their selectors see
		// the slice disappear.
		setTimeout(() => deleteKeyed(this.sim.store, this.storeKey), 0);
	}

	isDisposed(): boolean {
		return this.disposed;
	}

	// Stat-weight reference stats (persisted with the settings; one shared
	// 'epRefStat' version counter).
	getRefStat(kind: 'dpsRefStat' | 'healRefStat' | 'tankRefStat'): Stat | undefined {
		return this.slice()[kind] as Stat | undefined;
	}

	setRefStat(kind: 'dpsRefStat' | 'healRefStat' | 'tankRefStat', stat: Stat | undefined) {
		if (this.slice()[kind] === stat) return;
		this.write({ [kind]: stat }, ['epRefStat']);
	}

	getSpecIcon(): string {
		return this.playerSpec.getIcon('medium');
	}

	getPlayerSpec(): PlayerSpec<SpecType> {
		return this.playerSpec;
	}

	getSpec(): SpecType {
		return this.getPlayerSpec().specID;
	}

	getPlayerClass(): PlayerClass<SpecClasses<SpecType>> {
		return this.playerClass;
	}

	getClass(): SpecClasses<SpecType> {
		return this.playerSpec.classID;
	}

	getClassColor(): string {
		return this.playerClass.hexColor;
	}

	canEnableTargetDummies(): boolean {
		const healingSpellClasses: Class[] = [Class.ClassDruid, Class.ClassPaladin, Class.ClassPriest, Class.ClassShaman];
		return healingSpellClasses.includes(this.getClass());
	}

	shouldEnableTargetDummies(): boolean {
		if (this.getPlayerSpec().isHealingSpec) {
			return true;
		}

		if (!this.itemSwapSettings.getEnableItemSwap()) {
			return false;
		}

		// Not comprehensive, add other relevant IDs here as needed.
		const healingProcTrinkets: number[] = [72898, 77969, 77204, 77989];

		return this.itemSwapSettings.getGear().hasTrinketFromOptions(healingProcTrinkets);
	}

	// TODO: Cata - Check this
	isSpec<T extends Spec>(specId: T): this is Player<T> {
		return (this.getSpec() as unknown) == specId;
	}

	isClass<T extends Class>(classId: T): this is Player<ClassSpecs<T>> {
		return (this.getClass() as unknown) == classId;
	}

	getParty(): Party | null {
		return this.party;
	}

	getRaid(): Raid | null {
		return this.raid;
	}

	// Returns this player's index within its party [0-4].
	getPartyIndex(): number {
		if (this.party == null) {
			throw new Error("Can't get party index for player without a party!");
		}

		return this.party.getPlayers().indexOf(this);
	}

	// Returns this player's index within its raid [0-24].
	getRaidIndex(): number {
		if (this.party == null) {
			throw new Error("Can't get raid index for player without a party!");
		}

		return this.party.getIndex() * MAX_PARTY_SIZE + this.getPartyIndex();
	}

	// This should only ever be called from party.
	setParty(newParty: Party | null) {
		if (newParty == null) {
			this.party = null;
			this.raid = null;
		} else {
			this.party = newParty;
			this.raid = newParty.raid;
		}
	}

	getOtherPartyMembers(): Array<Player<any>> {
		if (this.party == null) {
			return [];
		}

		return this.party.getPlayers().filter(player => player != null && player != this) as Array<Player<any>>;
	}

	// Returns all items that this player can wear in the given slot.
	getItems(slot: ItemSlot): Array<Item> {
		return this.sim.db.getItems(slot).filter(item => canEquipItem(item, this.playerSpec, slot));
	}

	// Returns all random suffixes that this player would be interested in for the given base item.
	getRandomSuffixes(item: Item): Array<ItemRandomSuffix> {
		return item.randomSuffixOptions
			.map(id => this.sim.db.getRandomSuffixById(id))
			.filter((suffix): suffix is ItemRandomSuffix => !!suffix && this.computeRandomSuffixEP(suffix) > 0);
	}

	// Returns all enchants that this player can wear in the given slot.
	getEnchants(slot: ItemSlot): Array<Enchant> {
		return this.sim.db.getEnchants(slot).filter(enchant => canEquipEnchant(enchant, this.playerSpec, this.hasProfession(Profession.Enchanting)));
	}

	// Returns all gems that this player can wear of the given color.
	getGems(socketColor?: GemColor): Array<Gem> {
		return this.sim.db.getGems(socketColor);
	}

	getEpWeights(): Stats {
		return this.slice().epWeights;
	}

	setEpWeights(newEpWeights: Stats) {
		this.patch('epWeights', newEpWeights);

		this.gemEPCache = new Map();
		this.enchantEPCache = new Map();
		this.randomSuffixEPCache = new Map();
		for (let i = 0; i < ItemSlot.ItemSlotRanged + 1; ++i) {
			this.itemEPCache[i] = new Map();
		}
	}

	getDefaultEpRatios(isTankSpec: boolean, isHealingSpec: boolean): Array<number> {
		const defaultRatios = new Array(Player.numEpRatios).fill(0);
		if (this.specConfig?.epRatios) {
			if (this.specConfig.epRatios.length != Player.numEpRatios) {
				throw new Error(
					`Invalid number of EP ratios in spec config for spec ${this.getSpec()}. Expected ${Player.numEpRatios}, got ${this.specConfig.epRatios.length}`,
				);
			}
			this.specConfig.epRatios.forEach((ratio, index) => (defaultRatios[index] = ratio));
			return defaultRatios;
		}
		if (isHealingSpec) {
			// By default only value HPS EP for healing spec
			defaultRatios[1] = 1;
		} else if (isTankSpec) {
			// By default value TPS and DTPS EP equally for tanking spec
			defaultRatios[2] = 1;
			defaultRatios[3] = 1;
		} else {
			// By default only value DPS EP
			defaultRatios[0] = 1;
		}

		return defaultRatios;
	}

	getEpRatios() {
		return this.slice().epRatios.slice();
	}

	setEpRatios(newRatios: Array<number>) {
		this.patch('epRatios', newRatios);
	}

	hasCustomEPWeights(): boolean {
		return !this.getSpecConfig().presets.epWeights.some(epw => epw.epWeights.equals(this.getEpWeights()));
	}

	// Error display (toasts) is the caller's responsibility; a result with
	// result.error set or a thrown error must be handled by the UI layer.
	async computeStatWeights(
		epStats: Array<Stat>,
		epPseudoStats: Array<PseudoStat>,
		epReferenceStat: Stat,
		onProgress: WorkerProgressCallback,
	): Promise<StatWeightsResult> {
		return await this.sim.statWeights(this, epStats, epPseudoStats, epReferenceStat, onProgress);
	}

	getCurrentStats(): PlayerStats {
		return PlayerStats.clone(this.slice().currentStats);
	}

	setCurrentStats(newStats: PlayerStats) {
		this.patch('currentStats', newStats);
	}

	getMetadata(): UnitMetadata {
		return this.metadata;
	}

	getPetMetadatas(): UnitMetadataList {
		return this.petMetadatas;
	}

	async updateMetadata(): Promise<boolean> {
		const currentStats = this.slice().currentStats;
		const playerPromise = this.metadata.update(currentStats.metadata!);
		const petsPromise = this.petMetadatas.update(currentStats.pets.map(p => p.metadata!));
		const playerUpdated = await playerPromise;
		const petsUpdated = await petsPromise;
		return playerUpdated || petsUpdated;
	}

	getName(): string {
		return this.slice().name;
	}
	setName(newName: string) {
		if (newName != this.getName()) {
			this.patch('name', newName);
		}
	}

	getLabel(): string {
		if (this.party) {
			return `${this.getName()} (#${this.getRaidIndex() + 1})`;
		} else {
			return this.getName();
		}
	}

	getRace(): Race {
		return this.slice().race;
	}
	setRace(newRace: Race) {
		if (newRace != this.getRace()) {
			this.patch('race', newRace);
		}
	}

	getProfession1(): Profession {
		return this.slice().profession1 as Profession;
	}
	setProfession1(newProfession: Profession) {
		if (newProfession != this.getProfession1()) {
			this.patch('profession1', newProfession);
		}
	}
	getProfession2(): Profession {
		return this.slice().profession2 as Profession;
	}
	setProfession2(newProfession: Profession) {
		if (newProfession != this.getProfession2()) {
			this.patch('profession2', newProfession);
		}
	}
	getProfessions(): Array<Profession> {
		return [this.getProfession1(), this.getProfession2()].filter(p => p != Profession.ProfessionUnknown);
	}
	setProfessions(newProfessions: Array<Profession>) {
		batch(() => {
			this.setProfession1(newProfessions[0] || Profession.ProfessionUnknown);
			this.setProfession2(newProfessions[1] || Profession.ProfessionUnknown);
		});
	}
	hasProfession(prof: Profession): boolean {
		return this.getProfessions().includes(prof);
	}

	getFaction(): Faction {
		return raceToFaction[this.getRace()];
	}

	getBuffs(): IndividualBuffs {
		// Make a defensive copy
		return IndividualBuffs.clone(this.slice().buffs);
	}

	setBuffs(newBuffs: IndividualBuffs) {
		if (IndividualBuffs.equals(this.slice().buffs, newBuffs)) return;

		// Make a defensive copy
		this.patch('buffs', IndividualBuffs.clone(newBuffs));
	}

	getConsumes(forSimming?: boolean): ConsumesSpec {
		if (forSimming) {
			const epStats = [...(this.specConfig.consumableStats ?? []), ...this.specConfig.epStats];
			const dbPotions = this.sim.db.getConsumablesByTypeAndStats(ConsumableType.ConsumableTypePotion, epStats);
			const dbConjured = relevantConsumableOptions(CONJURED_CONFIG, this.specConfig)
				.filter(option => (option.showWhen ? option.showWhen(this) : true))
				.map(option => option.value);
			return ConsumesSpec.create({
				...this.slice().consumables,
				potions: dbPotions.map(p => p.id),
				conjuredItems: dbConjured,
			});
		}
		// Make a defensive copy
		return ConsumesSpec.clone(this.slice().consumables);
	}

	// Weapon stones grant their crit rating to a melee weapon only, but the back-end tracks a
	// single physical crit rating stat shared by melee and ranged, so ranged stat displays have to
	// offset them back out.
	getRangedImbueStatOffsets(): Stats {
		const isWeaponStone = (imbueId: number) => imbueId === ADAMANTITE_SHARPENING_STONE_ID || imbueId === ADAMANTITE_WEIGHTSTONE_ID;
		const consumables = this.slice().consumables;
		const party = this.getParty();
		const mhImbueApplied = !party || party.getBuffs().windfuryTotem === TristateEffect.TristateEffectMissing;

		let offsets = new Stats();
		if (mhImbueApplied && isWeaponStone(consumables.mhImbueId)) {
			offsets = offsets.addStat(Stat.StatMeleeCritRating, -14);
		}
		if (isWeaponStone(consumables.ohImbueId)) {
			offsets = offsets.addStat(Stat.StatMeleeCritRating, -14);
		}
		return offsets;
	}

	setConsumes(newConsumes: ConsumesSpec) {
		if (ConsumesSpec.equals(this.slice().consumables, newConsumes)) return;

		// Make a defensive copy
		this.patch('consumables', ConsumesSpec.clone(newConsumes));
	}

	equipItem(slot: ItemSlot, newItem: EquippedItem | null) {
		this.setGear(this.getGear().withEquippedItem(slot, newItem));
	}

	getEquippedItem(slot: ItemSlot): EquippedItem | null {
		return this.getGear().getEquippedItem(slot);
	}

	getEquippedItems(): Array<EquippedItem | null> {
		return this.getGear().getEquippedItems();
	}

	getGear(): Gear {
		return this.slice().gear;
	}

	setGear(newGear: Gear, forceUpdate?: boolean) {
		if (newGear.equals(this.getGear()) && !forceUpdate) return;

		// Weapon stone imbues are corrected in the same write as the gear, so that a subscriber
		// (including pickers that auto-clear a now-invalid selection) never sees the pair
		// disagree.
		const adjustedConsumes = newGear.adjustImbues(this.slice().consumables);
		if (adjustedConsumes !== this.slice().consumables) {
			this.write({ gear: newGear, consumables: adjustedConsumes }, ['gear', 'consumables']);
		} else {
			this.patch('gear', newGear);
		}
	}

	async setGearAsync(newGear: Gear, forceUpdate?: boolean) {
		if (newGear.equals(this.getGear()) && !forceUpdate) return;
		const statsUpdatePromise = new Promise<void>(resolve => {
			const unsub = this.sim.store.subscribe(
				s => s.players[this.storeKey]?.v.currentStats,
				() => {
					unsub();
					resolve();
				},
			);
		});
		this.setGear(newGear);
		await statsUpdatePromise;
	}

	getBonusStats(): Stats {
		return this.slice().bonusStats;
	}

	setBonusStats(newBonusStats: Stats) {
		if (newBonusStats.equals(this.getBonusStats())) return;

		this.patch('bonusStats', newBonusStats);
	}

	// The raid debuffs that the character sheet attributes as their own stage. Not one of the
	// server's cumulative stat stages, so it is derived here and fed to computeStatAttribution.
	getDebuffStats(): Stats {
		const debuffs = this.sim.raid.getDebuffs();
		const ownAgility = this.getCurrentStats().finalStats?.stats[Stat.StatAgility] ?? 0;
		return characterSheetDebuffStats(debuffs, characterSheetExposeWeaknessAgility(debuffs, this.getClass(), this.getTalentsString(), ownAgility));
	}

	getCritImmunityInfo() {
		const critImmuneCap = 5.6;
		const currentStats = this.slice().currentStats;
		const defense = currentStats.finalStats?.stats[Stat.StatDefenseRating] || 0;
		const resilience = currentStats.finalStats?.stats[Stat.StatResilienceRating] || 0;

		const defenseContribution = Math.floor(defense / Mechanics.DEFENSE_RATING_PER_DEFENSE_LEVEL) * Mechanics.MISS_DODGE_PARRY_BLOCK_CRIT_CHANCE_PER_DEFENSE;
		const resilienceContribution = resilience / Mechanics.RESILIENCE_RATING_PER_CRIT_REDUCTION_CHANCE;
		// PseudoStatReducedCritTakenPercent includes all sources: defense, resilience, and talents.
		const total = currentStats.finalStats?.pseudoStats[PseudoStat.PseudoStatReducedCritTakenPercent] || 0;
		const talentContribution = total - defenseContribution - resilienceContribution;

		return {
			total: total,
			delta: critImmuneCap - total,
			defense: defenseContribution,
			resilience: resilienceContribution,
			talents: talentContribution,
		};
	}

	getCritImmunity() {
		return this.getCritImmunityInfo().delta;
	}

	getMissChanceInfo() {
		const defense = this.slice().currentStats.finalStats?.stats[Stat.StatDefenseRating] || 0;
		const defenseContribution = Math.floor(defense / Mechanics.DEFENSE_RATING_PER_DEFENSE_LEVEL) * Mechanics.MISS_DODGE_PARRY_BLOCK_CRIT_CHANCE_PER_DEFENSE;
		let debuffs = 0;
		if (this.sim.raid.getDebuffs().scorpidSting) {
			debuffs = 5;
		} else if (this.sim.raid.getDebuffs().insectSwarm) {
			debuffs = 2;
		}

		return {
			base: 5,
			defense: defenseContribution,
			debuffs,
			total: 5 + defenseContribution + debuffs,
		};
	}

	getAvoidanceInfo() {
		const miss = this.getMissChanceInfo().total;
		const currentStats = this.slice().currentStats;
		const dodge = currentStats.finalStats?.pseudoStats[PseudoStat.PseudoStatDodgePercent] || 0;
		const parry = currentStats.finalStats?.pseudoStats[PseudoStat.PseudoStatParryPercent] || 0;
		let block = currentStats.finalStats?.pseudoStats[PseudoStat.PseudoStatBlockPercent] || 0;

		if (this.isSpec(Spec.SpecProtectionPaladin)) {
			block += 30;

			if (this.getEquippedItem(ItemSlot.ItemSlotRanged)?.id === 29388) {
				block += 42 / Mechanics.BLOCK_RATING_PER_BLOCK_PERCENT;
			}
		}

		return {
			miss: miss,
			dodge: dodge,
			parry: parry,
			block: block,
			total: miss + dodge + parry + block,
			shear: dodge + parry + block,
		};
	}

	getMeleeCritCapInfo(): MeleeCritCapInfo {
		const currentStats = this.slice().currentStats;
		const debuffStats = this.getDebuffStats();
		const debuffHit = debuffStats.getPseudoStat(PseudoStat.PseudoStatMeleeHitPercent) || 0;
		const debuffCrit = debuffStats.getPseudoStat(PseudoStat.PseudoStatMeleeCritPercent) || 0;

		const meleeCrit = (currentStats.finalStats?.pseudoStats[PseudoStat.PseudoStatMeleeCritPercent] || 0) + debuffCrit;
		const meleeHit = (currentStats.finalStats?.pseudoStats[PseudoStat.PseudoStatMeleeHitPercent] || 0) + debuffHit;
		const expertise = (currentStats.finalStats?.stats[Stat.StatExpertiseRating] || 0) / Mechanics.EXPERTISE_PER_QUARTER_PERCENT_REDUCTION / 4;
		const targetLevel = this.sim.encounter.primaryTarget.level;
		const critSuppression = { 68: 0, 70: 0, 71: 1, 72: 2, 73: 4.8 }[targetLevel] ?? 0;
		const hitSuppression = { 68: 0, 70: 0, 71: 0, 72: 0, 73: 1 }[targetLevel] ?? 0;
		const glancing = { 68: 0, 70: 6, 71: 12, 72: 18, 73: 24 }[targetLevel] ?? 0;

		const oneHandHitCap = ({ 68: 4, 70: 5, 71: 6, 72: 7, 73: 8 }[targetLevel] ?? 4) + hitSuppression;
		// DW Penalty is a fixed 19%
		const dualWieldHitCap = oneHandHitCap + 19;
		const hasOffhandWeapon = this.getGear().getEquippedItem(ItemSlot.ItemSlotOffHand)?.item.weaponSpeed !== undefined;
		// Due to warrior HS bug, hit cap for crit cap calculation should be 8% instead of 27%
		const meleeHitCap = hasOffhandWeapon && this.getClass() != Class.ClassWarrior ? dualWieldHitCap : oneHandHitCap;
		const dodgeCap = { 68: 4, 70: 5, 71: 5.5, 72: 6, 73: 6.5 }[targetLevel] ?? 4;
		const parryCap = this.getInFrontOfTarget() ? ({ 68: 4, 70: 5, 71: 5.5, 72: 6, 73: 14 }[targetLevel] ?? 4) : 0;
		const expertiseCap = dodgeCap + parryCap;

		const remainingMeleeHitCap = Math.max(meleeHitCap - meleeHit, 0.0);
		const remainingDodgeCap = Math.max(dodgeCap - expertise, 0.0);
		const remainingParryCap = Math.max(parryCap - expertise, 0.0);
		const remainingExpertiseCap = remainingDodgeCap + remainingParryCap;

		let specSpecificOffset = 0.0;
		if (this.getSpec() === Spec.SpecEnhancementShaman) {
			const player = this as unknown as Player<Spec.SpecEnhancementShaman>;
			// Elemental Devastation uptime is near 100%
			const ranks = player.getTalents().elementalDevastation;
			specSpecificOffset = 3.0 * ranks;
		}

		const baseCritCap = 100.0 - glancing + critSuppression - remainingMeleeHitCap - remainingExpertiseCap - specSpecificOffset;
		const playerCritCapDelta = meleeCrit - baseCritCap;

		return {
			meleeCrit,
			meleeHit,
			expertise,
			suppression: critSuppression,
			glancing,
			debuffCrit,
			hasOffhandWeapon,
			meleeHitCap,
			expertiseCap,
			remainingMeleeHitCap,
			remainingExpertiseCap,
			baseCritCap,
			specSpecificOffset,
			playerCritCapDelta,
		};
	}

	getMeleeCritCap() {
		return this.getMeleeCritCapInfo().playerCritCapDelta;
	}

	// In-place rotation mutation with a single change event — the write path for
	// the APL editor's per-field edits.
	modifyAplRotation(modify: (rotation: APLRotation) => void) {
		modify(this.aplRotation_);
		this.write({}, ['rotation']);
	}

	setAplRotation(newRotation: APLRotation) {
		if (APLRotation.equals(newRotation, this.aplRotation_)) return;

		this.aplRotation_ = APLRotation.clone(newRotation);
		this.write({}, ['rotation']);
	}

	getSimpleRotation(): SpecRotation<SpecType> {
		const jsonStr = this.aplRotation_.simple?.specRotationJson || '';
		if (!jsonStr) {
			return this.specTypeFunctions.rotationCreate();
		}

		try {
			const json = JSON.parse(jsonStr);
			return this.specTypeFunctions.rotationFromJson(json);
		} catch (e) {
			console.warn(`Error parsing rotation spec options: ${e}\n\nSpec options: '${jsonStr}'`);
			return this.specTypeFunctions.rotationCreate();
		}
	}

	setSimpleRotation(newRotation: SpecRotation<SpecType>) {
		if (this.specTypeFunctions.rotationEquals(newRotation, this.getSimpleRotation())) return;

		if (!this.aplRotation_.simple) {
			this.aplRotation_.simple = SimpleRotation.create();
		}
		this.aplRotation_.simple.specRotationJson = JSON.stringify(this.specTypeFunctions.rotationToJson(newRotation));

		this.write({}, ['rotation']);
	}

	getSimpleCooldowns(): Cooldowns {
		// Make a defensive copy
		return Cooldowns.clone(this.aplRotation_.simple?.cooldowns || Cooldowns.create());
	}

	setSimpleCooldowns(newCooldowns: Cooldowns) {
		if (Cooldowns.equals(this.getSimpleCooldowns(), newCooldowns)) return;

		if (!this.aplRotation_.simple) {
			this.aplRotation_.simple = SimpleRotation.create();
		}
		this.aplRotation_.simple.cooldowns = newCooldowns;
		this.write({}, ['rotation']);
	}

	getRotationType(): APLRotationType {
		if (this.aplRotation_.type == APLRotationType.TypeUnknown) {
			return APLRotationType.TypeAPL;
		} else {
			return this.aplRotation_.type;
		}
	}

	hasSimpleRotationGenerator(): boolean {
		return this.simpleRotationGenerator != null;
	}

	getResolvedAplRotation(forSimming?: boolean): APLRotation {
		const type = this.getRotationType();
		if (type == APLRotationType.TypeAuto && this.autoRotationGenerator) {
			// Clone to avoid modifying preset rotations, which are often returned directly.
			const rot = APLRotation.clone(this.autoRotationGenerator(this));
			rot.type = APLRotationType.TypeAuto;
			return rot;
		} else if (type == APLRotationType.TypeSimple && this.simpleRotationGenerator) {
			// Clone to avoid modifying preset rotations, which are often returned directly.
			const simpleRot = this.getSimpleRotation();
			const rot = APLRotation.clone(this.simpleRotationGenerator(this, simpleRot, this.getSimpleCooldowns()));
			rot.simple = this.aplRotation_.simple;
			rot.type = APLRotationType.TypeSimple;
			return rot;
		} else if (forSimming) {
			return this.aplRotation_;
		} else {
			return omitDeep(this.aplRotation_, ['uuid']);
		}
	}

	getTalents(): SpecTalents<SpecType> {
		if (this.talents == null) {
			this.talents = playerTalentStringToProto(this.playerSpec, this.getTalentsString()) as SpecTalents<SpecType>;
		}
		return this.talents!;
	}

	getTalentsString(): string {
		return this.slice().talentsString;
	}

	setTalentsString(newTalentsString: string) {
		if (newTalentsString == this.getTalentsString()) return;

		// Invalidate the parsed-talents cache before the emit fires.
		this.talents = null;
		this.patch('talentsString', newTalentsString);
	}

	getTalentTreePoints(): Array<number> {
		return getTalentTreePoints(this.getTalentsString());
	}

	getTalentTreeIcon(): string {
		return this.playerSpec.getIcon('medium');
	}

	getClassOptions(): ClassOptions<SpecType> {
		return this.getSpecOptions().classOptions as ClassOptions<SpecType>;
	}

	setClassOptions(newClassOptions: ClassOptions<SpecType>) {
		const newSpecOptions = this.getSpecOptions();
		newSpecOptions.classOptions = newClassOptions;
		if (this.specTypeFunctions.optionsEquals(newSpecOptions, this.slice().specOptions as SpecOptions<SpecType>)) return;

		this.patch('specOptions', this.specTypeFunctions.optionsCopy(newSpecOptions));
	}

	getSpecOptions(): SpecOptions<SpecType> {
		return this.specTypeFunctions.optionsCopy(this.slice().specOptions as SpecOptions<SpecType>);
	}

	setSpecOptions(newSpecOptions: SpecOptions<SpecType>) {
		if (this.specTypeFunctions.optionsEquals(newSpecOptions, this.slice().specOptions as SpecOptions<SpecType>)) return;

		this.patch('specOptions', this.specTypeFunctions.optionsCopy(newSpecOptions));
	}

	getReactionTime(): number {
		return this.slice().reactionTime;
	}

	setReactionTime(newReactionTime: number) {
		if (newReactionTime == this.getReactionTime()) return;

		this.patch('reactionTime', newReactionTime);
	}

	getChannelClipDelay(): number {
		return this.slice().channelClipDelay;
	}

	setChannelClipDelay(newChannelClipDelay: number) {
		if (newChannelClipDelay == this.getChannelClipDelay()) return;

		this.patch('channelClipDelay', newChannelClipDelay);
	}

	getInFrontOfTarget(): boolean {
		return this.slice().inFrontOfTarget;
	}

	setInFrontOfTarget(newInFrontOfTarget: boolean) {
		if (newInFrontOfTarget == this.getInFrontOfTarget()) return;

		this.patch('inFrontOfTarget', newInFrontOfTarget);
	}

	getDistanceFromTarget(): number {
		return this.slice().distanceFromTarget;
	}

	setDistanceFromTarget(newDistanceFromTarget: number) {
		if (newDistanceFromTarget == this.getDistanceFromTarget()) return;

		this.patch('distanceFromTarget', newDistanceFromTarget);
	}

	setDefaultHealingParams(hm: HealingModel) {
		const boss = this.sim.encounter.primaryTarget;
		const dualWield = boss.dualWield;
		if (hm.cadenceSeconds == 0) {
			let maxCadence = 1.5 * boss.swingSpeed;
			if (dualWield) {
				maxCadence /= 2;
			}
			hm.cadenceSeconds = 0.4;
			hm.cadenceVariation = maxCadence - hm.cadenceSeconds;
		}
		if (hm.hps == 0) {
			hm.hps = (0.25 * boss.minBaseDamage) / boss.swingSpeed;
			if (dualWield) {
				hm.hps *= 1.5;
			}
		}
	}

	enableHealing() {
		this.healingEnabled = true;
		const hm = this.getHealingModel();
		if (hm.cadenceSeconds == 0 || hm.hps == 0) {
			this.setDefaultHealingParams(hm);
			this.setHealingModel(hm);
		}
	}

	getHealingModel(): HealingModel {
		// Make a defensive copy
		return HealingModel.clone(this.slice().healingModel);
	}

	setHealingModel(newHealingModel: HealingModel) {
		if (HealingModel.equals(this.slice().healingModel, newHealingModel)) return;

		// Make a defensive copy
		const healingModel = HealingModel.clone(newHealingModel);
		// If we have enabled healing model and try to set 0s cadence or 0 incoming HPS, then set intelligent defaults instead based on boss parameters.
		if (this.healingEnabled) {
			this.setDefaultHealingParams(healingModel);
		}
		this.patch('healingModel', healingModel);
	}

	computeStatsEP(stats?: Stats): number {
		if (stats == undefined) {
			return 0;
		}
		return stats.computeEP(this.getEpWeights());
	}

	computeGemEP(gem: Gem): number {
		if (this.gemEPCache.has(gem.id)) {
			return this.gemEPCache.get(gem.id)!;
		}

		const epFromStats = this.computeStatsEP(new Stats(gem.stats));
		const epFromEffect = getMetaGemEffectEP(this.playerSpec, gem, this.getEpWeights());
		let bonusEP = 0;
		// unique items are slightly worse than non-unique because you can have only one.
		if (gem.unique) {
			bonusEP -= 0.01;
		}

		const ep = epFromStats + epFromEffect + bonusEP;
		this.gemEPCache.set(gem.id, ep);
		return ep;
	}

	computeEnchantEP(enchant: Enchant): number {
		if (this.enchantEPCache.has(enchant.effectId)) {
			return this.enchantEPCache.get(enchant.effectId)!;
		}

		const ep = this.computeStatsEP(new Stats(enchant.stats));
		this.enchantEPCache.set(enchant.effectId, ep);
		return ep;
	}

	computeRandomSuffixEP(randomSuffix: ItemRandomSuffix): number {
		if (this.randomSuffixEPCache.has(randomSuffix.id)) {
			return this.randomSuffixEPCache.get(randomSuffix.id)!;
		}

		const ep = this.computeStatsEP(new Stats(randomSuffix.stats));
		this.randomSuffixEPCache.set(randomSuffix.id, ep);
		return ep;
	}

	computeItemEP(item: Item, slot: ItemSlot): number {
		if (item == null) return 0;

		const cacheKey = `${item.id}-${JSON.stringify(this.getEpWeights())}`;

		const cached = this.itemEPCache[slot].get(cacheKey);
		if (cached !== undefined) return cached;

		const equippedItem = new EquippedItem({ item }).withDynamicStats();
		const itemStats = equippedItem.calcStats(slot);

		// For random suffix items, use the suffix option with the highest EP for the purposes of ranking items in the picker.
		let maxSuffixEP = 0;
		if (item.randomSuffixOptions.length) {
			const suffixEPs = equippedItem.item.randomSuffixOptions.map(id => this.computeRandomSuffixEP(this.sim.db.getRandomSuffixById(id)! || 0));
			maxSuffixEP = (Math.max(...suffixEPs) * equippedItem.item.randPropPoints) / 10000;
		}

		let ep = itemStats.computeEP(this.getEpWeights()) + maxSuffixEP;

		// unique items are slightly worse than non-unique because you can have only one.
		if (item.unique) {
			ep -= 0.01;
		}

		// Compare whether its better to match sockets + get socket bonus, or just use best gems.
		const bestGemEPNotMatchingSockets = sum(
			item.gemSockets.map(socketColor => {
				const gems = this.sim.db.getGems(socketColor).filter(gem => isUnrestrictedGem(gem, this.sim.getPhase()));
				if (gems.length > 0) {
					return Math.max(...gems.map(gem => this.computeGemEP(gem)));
				} else {
					return 0;
				}
			}),
		);

		const bestGemEPMatchingSockets =
			sum(
				item.gemSockets.map(socketColor => {
					const gems = this.sim.db
						.getGems(socketColor)
						.filter(gem => isUnrestrictedGem(gem, this.sim.getPhase()) && gemMatchesSocket(gem, socketColor));
					if (gems.length > 0) {
						return Math.max(...gems.map(gem => this.computeGemEP(gem)));
					} else {
						return 0;
					}
				}),
			) + this.computeStatsEP(new Stats(item.socketBonus));

		ep += Math.max(bestGemEPMatchingSockets, bestGemEPNotMatchingSockets);

		this.itemEPCache[slot].set(cacheKey, ep);
		return ep;
	}

	static ARMOR_SLOTS: Array<ItemSlot> = [
		ItemSlot.ItemSlotHead,
		ItemSlot.ItemSlotShoulder,
		ItemSlot.ItemSlotChest,
		ItemSlot.ItemSlotWrist,
		ItemSlot.ItemSlotHands,
		ItemSlot.ItemSlotLegs,
		ItemSlot.ItemSlotWaist,
		ItemSlot.ItemSlotFeet,
	];

	static WEAPON_SLOTS: Array<ItemSlot> = [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand];

	static readonly DIFFICULTY_SRCS: Partial<Record<SourceFilterOption, DungeonDifficulty>> = {
		[SourceFilterOption.SourceDungeon]: DungeonDifficulty.DifficultyNormal,
		[SourceFilterOption.SourceDungeonH]: DungeonDifficulty.DifficultyHeroic,
		[SourceFilterOption.SourceRaidRF]: DungeonDifficulty.DifficultyRaid25RF,
		[SourceFilterOption.SourceRaid]: DungeonDifficulty.DifficultyRaid25,
		[SourceFilterOption.SourceRaidH]: DungeonDifficulty.DifficultyRaid25H,
		[SourceFilterOption.SourceRaidFlex]: DungeonDifficulty.DifficultyRaidFlex,
	};

	static readonly HEROIC_TO_NORMAL: Partial<Record<DungeonDifficulty, DungeonDifficulty>> = {
		[DungeonDifficulty.DifficultyHeroic]: DungeonDifficulty.DifficultyNormal,
		[DungeonDifficulty.DifficultyRaid10H]: DungeonDifficulty.DifficultyRaid10,
		[DungeonDifficulty.DifficultyRaid25H]: DungeonDifficulty.DifficultyRaid25,
	};

	static readonly RAID_IDS: Partial<Record<RaidFilterOption, number>> = {
		[RaidFilterOption.RaidKara]: 3457,
		[RaidFilterOption.RaidGruul]: 3923,
		[RaidFilterOption.RaidMag]: 3836,
		[RaidFilterOption.RaidTK]: 3845,
		[RaidFilterOption.RaidSSC]: 3607,
		[RaidFilterOption.RaidMH]: 3606,
		[RaidFilterOption.RaidBT]: 3959,
		[RaidFilterOption.RaidZA]: 3805,
		[RaidFilterOption.RaidSWP]: 4075,
	};

	filterItemData<T>(itemData: Array<T>, getItemFunc: (val: T) => Item, slot: ItemSlot): Array<T> {
		const filters = this.sim.getFilters();

		const filterItems = (itemData: Array<T>, filterFunc: (item: Item) => boolean) => {
			return itemData.filter(itemElem => filterFunc(getItemFunc(itemElem)));
		};

		if (filters.minIlvl != 0) {
			itemData = filterItems(itemData, item => (item.scalingOptions?.[0].ilvl || item.ilvl) >= filters.minIlvl);
		}
		if (filters.maxIlvl != 0) {
			itemData = filterItems(itemData, item => (item.scalingOptions?.[0].ilvl || item.ilvl) <= filters.maxIlvl);
		}

		if (filters.factionRestriction != UIItem_FactionRestriction.UNSPECIFIED) {
			itemData = filterItems(
				itemData,
				item => item.factionRestriction == filters.factionRestriction || item.factionRestriction == UIItem_FactionRestriction.UNSPECIFIED,
			);
		}

		if (!filters.sources.includes(SourceFilterOption.SourceSoldBy)) {
			itemData = filterItems(itemData, item => !item.sources.some(itemSrc => itemSrc.source.oneofKind == 'soldBy'));
		}
		if (!filters.sources.includes(SourceFilterOption.SourceCrafting)) {
			itemData = filterItems(itemData, item => !item.sources.some(itemSrc => itemSrc.source.oneofKind == 'crafted'));
		}
		if (!filters.sources.includes(SourceFilterOption.SourceQuest)) {
			itemData = filterItems(itemData, item => !item.sources.some(itemSrc => itemSrc.source.oneofKind == 'quest'));
		}
		if (!filters.sources.includes(SourceFilterOption.SourceReputation)) {
			itemData = filterItems(itemData, item => !item.sources.some(itemSrc => itemSrc.source.oneofKind == 'rep'));
		}
		if (!filters.sources.includes(SourceFilterOption.SourcePvp)) {
			itemData = filterItems(itemData, item => !isPVPItem(item));
		}

		for (const [srcOptionStr, difficulty] of Object.entries(Player.DIFFICULTY_SRCS)) {
			const srcOption = parseInt(srcOptionStr) as SourceFilterOption;

			if (!filters.sources.includes(srcOption)) {
				itemData = filterItems(
					itemData,
					item => !item.sources.some(itemSrc => itemSrc.source.oneofKind == 'drop' && itemSrc.source.drop.difficulty == difficulty),
				);

				if (difficulty == DungeonDifficulty.DifficultyRaid10H || difficulty == DungeonDifficulty.DifficultyRaid25H) {
					const normalDifficulty = Player.HEROIC_TO_NORMAL[difficulty];
					itemData = filterItems(
						itemData,
						item =>
							!item.sources.some(
								itemSrc =>
									itemSrc.source.oneofKind == 'drop' &&
									itemSrc.source.drop.difficulty == normalDifficulty &&
									itemSrc.source.drop.category == AL_CATEGORY_HARD_MODE,
							),
					);
				}
			}
		}

		for (const [raidOptionStr, zoneId] of Object.entries(Player.RAID_IDS)) {
			const raidOption = parseInt(raidOptionStr) as RaidFilterOption;
			if (!filters.raids.includes(raidOption)) {
				itemData = filterItems(
					itemData,
					item => !item.sources.some(itemSrc => itemSrc.source.oneofKind == 'drop' && itemSrc.source.drop.zoneId == zoneId),
				);
			}
		}

		if (Player.ARMOR_SLOTS.includes(slot)) {
			itemData = filterItems(itemData, item => {
				if (!filters.armorTypes.includes(item.armorType)) {
					return false;
				}

				return true;
			});
		} else if (Player.WEAPON_SLOTS.includes(slot)) {
			itemData = filterItems(itemData, item => {
				if (!filters.weaponTypes.includes(item.weaponType)) {
					return false;
				}
				if (!filters.oneHandedWeapons && item.handType != HandType.HandTypeTwoHand) {
					return false;
				}
				if (!filters.twoHandedWeapons && item.handType == HandType.HandTypeTwoHand) {
					return false;
				}

				const minSpeed = slot == ItemSlot.ItemSlotMainHand ? filters.minMhWeaponSpeed : filters.minOhWeaponSpeed;
				const maxSpeed = slot == ItemSlot.ItemSlotMainHand ? filters.maxMhWeaponSpeed : filters.maxOhWeaponSpeed;
				if (minSpeed > 0 && item.weaponSpeed < minSpeed) {
					return false;
				}
				if (maxSpeed > 0 && item.weaponSpeed > maxSpeed) {
					return false;
				}

				return true;
			});
		} else if (slot == ItemSlot.ItemSlotRanged) {
			itemData = filterItems(itemData, item => {
				if (!filters.rangedWeaponTypes.includes(item.rangedWeaponType)) {
					return false;
				}

				const minSpeed = filters.minRangedWeaponSpeed;
				const maxSpeed = filters.maxRangedWeaponSpeed;
				if (minSpeed > 0 && item.weaponSpeed < minSpeed) {
					return false;
				}
				if (maxSpeed > 0 && item.weaponSpeed > maxSpeed) {
					return false;
				}

				return true;
			});
		}

		return itemData;
	}

	filterEnchantData<T>(enchantData: Array<T>, getEnchantFunc: (val: T) => Enchant, slot: ItemSlot, currentEquippedItem: EquippedItem | null): Array<T> {
		if (!currentEquippedItem) {
			return enchantData;
		}

		//const filters = this.sim.getFilters();

		return enchantData.filter(enchantElem => {
			const enchant = getEnchantFunc(enchantElem);

			if (!enchantAppliesToItem(enchant, currentEquippedItem.item)) {
				return false;
			}

			return true;
		});
	}

	filterGemData<T>(gemData: Array<T>, getGemFunc: (val: T) => Gem, slot: ItemSlot, socketColor: GemColor): Array<T> {
		const filters = this.sim.getFilters();

		const isJewelcrafting = this.hasProfession(Profession.Jewelcrafting);
		return gemData.filter(gemElem => {
			const gem = getGemFunc(gemElem);
			if (!isJewelcrafting && gem.requiredProfession == Profession.Jewelcrafting) {
				return false;
			}

			if (filters.matchingGemsOnly && !gemMatchesSocket(gem, socketColor)) {
				return false;
			}

			// MoP drops gems whose positive stats are all outside the spec's EP/gem stat list. TBC
			// does not: a meta's effect is not in its stat block at all, and the per-spec
			// allowlist was found to hide gems people wanted. Every gem that reaches here is kept.
			return true;
		});
	}

	makeUnitReference(): UnitReference {
		if (this.party == null) {
			return emptyUnitReference();
		} else {
			return newUnitReference(this.getRaidIndex());
		}
	}

	private toDatabase(): SimDatabase {
		const dbGear = this.getGear().toDatabase(this.sim.db);
		const dbItemSwapGear = this.itemSwapSettings.getGear().toDatabase(this.sim.db);
		return Database.mergeSimDatabases(dbGear, dbItemSwapGear);
	}

	toProto(forExport?: boolean, forSimming?: boolean, exportCategories?: Array<SimSettingCategories>): PlayerProto {
		const exportCategory = (cat: SimSettingCategories) => !exportCategories || exportCategories.length == 0 || exportCategories.includes(cat);

		const gear = this.getGear();
		const aplRotation = forSimming ? this.getResolvedAplRotation(forSimming) : omitDeep(this.aplRotation_, ['uuid']);

		let player = PlayerProto.create({
			apiVersion: CURRENT_API_VERSION,
			class: this.getClass(),
			database: forExport ? undefined : this.toDatabase(),
		});
		if (exportCategory(SimSettingCategories.Gear)) {
			PlayerProto.mergePartial(player, {
				equipment: gear.asSpec(),
				bonusStats: this.getBonusStats().toProto(),
				enableItemSwap: this.itemSwapSettings.getEnableItemSwap(),
				itemSwap: this.itemSwapSettings.toProto(),
			});
		}
		if (exportCategory(SimSettingCategories.Talents)) {
			PlayerProto.mergePartial(player, {
				talentsString: this.getTalentsString(),
			});
		}
		if (exportCategory(SimSettingCategories.Rotation)) {
			PlayerProto.mergePartial(player, {
				cooldowns: Cooldowns.create({
					hpPercentForDefensives: this.getSimpleCooldowns().hpPercentForDefensives,
				}),
				rotation: aplRotation,
			});
		}
		if (exportCategory(SimSettingCategories.Consumes)) {
			PlayerProto.mergePartial(player, {
				consumables: this.getConsumes(forSimming),
			});
		}
		if (exportCategory(SimSettingCategories.Miscellaneous)) {
			PlayerProto.mergePartial(player, {
				name: this.getName(),
				race: this.getRace(),
				profession1: this.getProfession1(),
				profession2: this.getProfession2(),
				reactionTimeMs: this.getReactionTime(),
				channelClipDelayMs: this.getChannelClipDelay(),
				inFrontOfTarget: this.getInFrontOfTarget(),
				distanceFromTarget: this.getDistanceFromTarget(),
				healingModel: this.getHealingModel(),
			});
			player = withSpec(this.getSpec(), player, this.getSpecOptions());
		}
		if (exportCategory(SimSettingCategories.External)) {
			PlayerProto.mergePartial(player, {
				buffs: this.getBuffs(),
			});
		}
		return player;
	}

	fromProto(proto: PlayerProto, includeCategories?: Array<SimSettingCategories>) {
		// Fix potential out-of-date protos before importing
		batch(() => {
			Player.updateProtoVersion(proto);
			const loadCategory = (cat: SimSettingCategories) => !includeCategories || includeCategories.length == 0 || includeCategories.includes(cat);
			if (loadCategory(SimSettingCategories.Gear)) {
				this.setGear(proto.equipment ? this.sim.db.lookupEquipmentSpec(proto.equipment) : new Gear({}));
				this.itemSwapSettings.setItemSwapSettings(
					proto.enableItemSwap,
					proto.itemSwap ? this.sim.db.lookupItemSwap(proto.itemSwap) : new ItemSwapGear({}),
					Stats.fromProto(proto.itemSwap?.prepullBonusStats),
				);
				this.setBonusStats(Stats.fromProto(proto.bonusStats || UnitStats.create()));
				//this.setBulkEquipmentSpec(BulkEquipmentSpec.create()); // Do not persist the bulk equipment settings.
			}
			if (loadCategory(SimSettingCategories.Talents)) {
				this.setTalentsString(proto.talentsString);
			}
			if (loadCategory(SimSettingCategories.Rotation)) {
				if (proto.rotation?.type == APLRotationType.TypeUnknown) {
					if (!proto.rotation) {
						proto.rotation = APLRotation.create();
					}
					proto.rotation.type = APLRotationType.TypeAuto;
				}
				this.setAplRotation(proto.rotation || APLRotation.create());
			}
			if (loadCategory(SimSettingCategories.Consumes)) {
				this.setConsumes(proto.consumables || ConsumesSpec.create());
			}
			if (loadCategory(SimSettingCategories.Miscellaneous)) {
				this.setSpecOptions(this.specTypeFunctions.optionsFromPlayer(proto));
				this.setName(proto.name);
				this.setRace(proto.race);
				this.setProfession1(proto.profession1);
				this.setProfession2(proto.profession2);
				this.setReactionTime(proto.reactionTimeMs);
				this.setChannelClipDelay(proto.channelClipDelayMs);
				this.setInFrontOfTarget(proto.inFrontOfTarget);
				this.setDistanceFromTarget(proto.distanceFromTarget);
				this.setHealingModel(proto.healingModel || HealingModel.create());
			}
			if (loadCategory(SimSettingCategories.External)) {
				this.setBuffs(proto.buffs || IndividualBuffs.create());
			}
		});
	}

	clone(): Player<SpecType> {
		const newPlayer = new Player<SpecType>(this.playerSpec, this.sim);
		newPlayer.fromProto(this.toProto());
		return newPlayer;
	}

	applySharedDefaults() {
		batch(() => {
			this.setReactionTime(100);
			this.setInFrontOfTarget(this.playerSpec.isTankSpec);
			this.setHealingModel(
				HealingModel.create({
					burstWindow: this.playerSpec.isTankSpec ? 6 : 0,
				}),
			);
			this.setSimpleCooldowns(
				Cooldowns.create({
					hpPercentForDefensives: this.playerSpec.isTankSpec ? 0.4 : 0,
				}),
			);
			this.setBonusStats(new Stats());
		});
	}

	getBaseDefense(): number {
		return Mechanics.CHARACTER_LEVEL * 5;
	}

	static updateProtoVersion(playerProto: PlayerProto) {
		if (!(playerProto.apiVersion < CURRENT_API_VERSION)) {
			return;
		}

		const conversionMap: ProtoConversionMap<PlayerProto> = new Map([
			[
				12,
				(oldProto: PlayerProto) => {
					oldProto.apiVersion = 13;

					// v12: ret paladin useConsecrate(bool) -> consecrationRank(int32).
					if (playerProto.spec?.oneofKind === 'retributionPaladin') {
						const jsonStr = playerProto.rotation?.simple?.specRotationJson;
						if (jsonStr) {
							try {
								const parsed = JSON.parse(jsonStr);

								if (!parsed.aura) {
									parsed.aura = 'SanctityAura';
								}

								if (parsed.useConsecrate) {
									parsed.consecrationRank = 6;
								}

								delete parsed.useConsecrate;

								playerProto.rotation!.simple!.specRotationJson = JSON.stringify(parsed);
							} catch {
								// Malformed JSON - nothing to migrate.
							}
						}
					}

					return oldProto;
				},
			],
		]);

		// Run the migration utility using the above map.
		migrateOldProto<PlayerProto>(playerProto, playerProto.apiVersion, conversionMap);

		// Flag the version as up-to-date once all migrations are done.
		playerProto.apiVersion = CURRENT_API_VERSION;
	}

	getSpecConfig(): SpecConfigData<SpecType> {
		return this.specConfig;
	}

	// Returns true/false for main-hand / off-hand
	getActiveRacialExpertiseBonuses(): [boolean, boolean] {
		const mainHand = this.getEquippedItem(ItemSlot.ItemSlotMainHand);
		const offHand = this.getEquippedItem(ItemSlot.ItemSlotOffHand);

		if (!mainHand && !offHand) {
			return [false, false];
		}

		switch (this.getRace()) {
			case Race.RaceHuman:
				return [
					mainHand?.item.weaponType === WeaponType.WeaponTypeMace || mainHand?.item.weaponType === WeaponType.WeaponTypeSword,
					offHand?.item.weaponType === WeaponType.WeaponTypeMace || offHand?.item.weaponType === WeaponType.WeaponTypeSword,
				];
			case Race.RaceOrc:
				return [mainHand?.item.weaponType === WeaponType.WeaponTypeAxe, offHand?.item.weaponType === WeaponType.WeaponTypeAxe];
		}

		return [false, false];
	}
}
