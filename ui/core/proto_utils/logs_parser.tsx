import clsx from 'clsx';

import { CacheHandler } from '../cache_handler';
import { RaidSimResult } from '../proto/api.js';
import { SpellSchool } from '../proto/common';
import { ResourceType } from '../proto/spell';
import { bucket, getEnumValues, stringComparator, sum } from '../utils.js';
import { ActionId } from './action_id.js';
import { resourceNames, spellSchoolNames, stringToResourceType } from './names.js';

export class Entity {
	readonly name: string;
	readonly ownerName: string; // Blank if not a pet.

	// Either target index, player index, or owner index depending on what kind
	// of entity this is.
	readonly index: number;

	readonly isTarget: boolean;
	readonly isPet: boolean;

	constructor(name: string, ownerName: string, index: number, isTarget: boolean, isPet: boolean) {
		this.name = name;
		this.ownerName = ownerName;
		this.index = index;
		this.isTarget = isTarget;
		this.isPet = isPet;
	}

	equals(other: Entity) {
		return this.isTarget == other.isTarget && this.isPet == other.isPet && this.index == other.index && this.name == other.name;
	}

	toString(): string {
		if (this.isTarget) {
			return `Target ${this.index + 1}`;
		} else if (this.isPet) {
			return `${this.ownerName} (#${this.index + 1}) - ${this.name}`;
		} else {
			return `${this.name} (#${this.index + 1})`;
		}
	}

	toHTML() {
		if (this.isTarget) {
			return <span className="text-danger">[Target {this.index + 1}]</span>;
		} else if (this.isPet) {
			return (
				<>
					<span className="text-primary">{`[${this.ownerName} ${this.index + 1}]`}</span>
					{` - `}
					{this.name}
				</>
			);
		} else {
			return <span className="text-primary">{`[${this.name} ${this.index + 1}]`}</span>;
		}
	}

	// Parses one or more Entities from a string.
	// Each entity label should be one of:
	//   'Target 1' if a target,
	//   'PlayerName (#1)' if a player, or
	//   'PlayerName (#1) - PetName' if a pet.
	static parseRegex = /\[(Target (\d+))|(([a-zA-Z0-9]+) \(#(\d+)\) - ([a-zA-Z0-9\s:,]+))|(([a-zA-Z0-9\s]+) \(#(\d+)\))\]/g;
	static parseAll(str: string): Array<Entity> {
		return Array.from(str.matchAll(Entity.parseRegex)).map(match => {
			if (match[1]) {
				return new Entity(match[1], '', parseInt(match[2]) - 1, true, false);
			} else if (match[3]) {
				return new Entity(match[6], match[4], parseInt(match[5]) - 1, false, true);
			} else if (match[7]) {
				return new Entity(match[8], '', parseInt(match[9]) - 1, false, false);
			} else {
				throw new Error('Invalid Entity match');
			}
		});
	}
}

interface SimLogParams {
	raw: string;
	logIndex: number;
	timestamp: number;
	source: Entity | null;
	target: Entity | null;
	actionId: ActionId | null;
	spellSchool: SpellSchool | null;
	threat: number;
}

// Bounded: this holds detached DOM, and an unbounded CacheHandler never evicts.
const cachedActionIdLink = new CacheHandler<HTMLAnchorElement>({ keysToKeep: 512 });

export class SimLog {
	readonly raw: string;

	// Index of this log within the full log output.
	// When comparing timestamps this should be used instead of timestamp, because
	// timestamp is scraped from log text and doesn't have enough precision.
	readonly logIndex: number;

	// Time in seconds from the encounter start.
	readonly timestamp: number;

	readonly source: Entity | null;
	readonly target: Entity | null;
	readonly actionId: ActionId | null;
	readonly actionIdAsString: string | null = null;

	// Spell schoool from this event. Note that not all events have spell schools, so this will be 0null.
	readonly spellSchool: SpellSchool | null;
	// Amount of threat generated from this event. Note that not all events generate threat, so this will be 0.
	readonly threat: number;

	// Logs for auras that were active at this timestamp.
	// This is only filled if populateActiveAuras() is called.
	activeAuras: Array<AuraUptimeLog>;

	constructor(params: SimLogParams) {
		this.raw = params.raw;
		this.logIndex = params.logIndex;
		this.timestamp = params.timestamp;
		this.source = params.source;
		this.target = params.target;
		this.actionId = params.actionId;
		try {
			this.actionIdAsString = this.actionId?.toString() || null;
		} catch {}
		this.spellSchool = params.spellSchool;
		this.threat = params.threat;
		this.activeAuras = [];
	}

	toHTML(includeTimestamp = true) {
		let html = <>{this.raw}</>;
		// Base logs already have the timestamp appended by default
		if (!includeTimestamp) {
			// One exec: this used to match and then exec the same pattern over the same string.
			// It now runs per visible row rather than once per log for the life of the result.
			const captureArr = TIMESTAMP_PREFIX_REGEX.exec(this.raw);
			if (captureArr && captureArr.length == 4) {
				html = <>{captureArr[3]}</>;
			}
		}

		if (this.source) {
			html = (
				<>
					{this.source.toHTML()} {html}
				</>
			);
		}
		return html;
	}

	toPrefix(includeTimestamp = true) {
		let prefix = '';
		if (includeTimestamp) {
			prefix = `[${this.timestamp.toFixed(2)}]`;
		}

		return (
			<>
				{prefix}
				{this.source?.toHTML()}
			</>
		);
	}

	rawWithoutTimestamp(): string {
		const captureArr = TIMESTAMP_PREFIX_REGEX.exec(this.raw);
		if (!captureArr || captureArr.length != 4) return this.raw;
		return `${captureArr[2] ?? ''}${captureArr[3]}`.trim();
	}

	formattedTimestamp(): string {
		const positiveTimestamp = Math.abs(this.timestamp);
		const minutes = Math.floor(positiveTimestamp / 60);
		const seconds = Math.floor(positiveTimestamp - minutes * 60);
		const milliseconds = ((positiveTimestamp - Math.floor(positiveTimestamp)) * 1000).toFixed();

		let formatted = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}:${String(milliseconds).padStart(3, '0')}`;
		if (this.timestamp < 0) {
			formatted = `-${formatted}`;
		}
		return formatted;
	}

	protected newActionIdLink(isAura?: boolean) {
		const cacheKey = this.actionIdAsString ? `${this.actionIdAsString}${isAura || ''}` : undefined;
		const cachedLink = cacheKey ? cachedActionIdLink.get(cacheKey) : null;
		if (cachedLink) return cachedLink.cloneNode(true);

		const iconElem = (<span className="icon icon-sm"></span>) as HTMLSpanElement;
		const actionAnchor = (
			<a className="log-action" target="_blank">
				<span>
					{iconElem} {this.actionId!.name}
				</span>
			</a>
		) as HTMLAnchorElement;
		this.actionId?.setBackground(iconElem);
		this.actionId?.setWowheadHref(actionAnchor);
		const datasetSet = this.actionId?.setWowheadDataset(actionAnchor, { useBuffAura: isAura }) ?? Promise.resolve();
		if (cacheKey) {
			datasetSet.then(() => cachedActionIdLink.set(cacheKey, actionAnchor.cloneNode(true) as HTMLAnchorElement)).catch(() => {});
		}
		return actionAnchor;
	}

	static async parseAll(result: RaidSimResult): Promise<SimLog[]> {
		const lines = result.logs.split('\n');
		const pending: Array<PendingLog> = new Array(lines.length);
		// Resolving an ActionId is the only asynchronous part of parsing, and a fight names
		// on the order of a hundred of them across tens of thousands of lines. Collect the
		// distinct ones while classifying, resolve them in one batch, then build every log
		// synchronously.
		const distinctIds = new Map<string, ActionIdRequest>();

		for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
			let line = lines[lineIndex];
			const params: SimLogParams = {
				raw: line,
				logIndex: lineIndex,
				timestamp: 0,
				source: null,
				target: null,
				actionId: null,
				spellSchool: null,
				threat: 0,
			};
			const spellSchoolMatch = SPELL_SCHOOL_REGEX.exec(line);
			if (spellSchoolMatch) {
				params.spellSchool = parseInt(spellSchoolMatch[1]);
			}

			const threatMatch = THREAT_REGEX.exec(line);
			if (threatMatch) {
				params.threat = parseFloat(threatMatch[1]);
				line = line.substring(0, threatMatch.index);
			}

			const match = TIMESTAMP_REGEX.exec(line);
			if (!match || !match[1]) {
				pending[lineIndex] = { params, matcher: null, match: null, key: '' };
				continue;
			}

			params.timestamp = parseFloat(match[1]);
			const remainder = match[2];

			const entities = Entity.parseAll(remainder);
			params.source = entities[0] || null;
			params.target = entities[1] || null;

			const matched = matchLogLine(params.raw);
			if (!matched) {
				pending[lineIndex] = { params, matcher: null, match: null, key: '' };
				continue;
			}
			// Keep the key: the build pass needs it, and recomputing it there means a second
			// idString() call and a second string per matched line.
			const logString = matched.matcher.idString(matched.match);
			const key = `${params.source?.index ?? -1}|${logString}`;
			pending[lineIndex] = { params, ...matched, key };
			if (!distinctIds.has(key)) {
				distinctIds.set(key, { logString, playerIndex: params.source?.index });
			}
		}

		const keys = [...distinctIds.keys()];
		const resolved = await Promise.all(
			keys.map(key => {
				const { logString, playerIndex } = distinctIds.get(key)!;
				return ActionId.fromLogString(logString).fill(playerIndex);
			}),
		);
		// ActionId is immutable (readonly fields, private constructor), so one resolved
		// instance can be shared by every line that names it.
		const actionIds = new Map<string, ActionId>();
		keys.forEach((key, i) => actionIds.set(key, resolved[i]));

		const logs = pending.map(entry => {
			if (!entry.matcher) {
				return new SimLog(entry.params);
			}
			entry.params.actionId = actionIds.get(entry.key)!;
			return entry.matcher.build(entry.params, entry.match!);
		});

		await Promise.all(
			logs.map(async log => {
				if (log instanceof CastFailedLog) {
					log.reason = (await ActionId.replaceAllInString(log.reason)).replace(/[.!?]$/, '');
				}
			}),
		);

		return logs;
	}

	isDamageDealt(): this is DamageDealtLog {
		return this instanceof DamageDealtLog;
	}

	isResourceChanged(): this is ResourceChangedLog {
		return this instanceof ResourceChangedLog;
	}

	isAuraEvent(): this is AuraEventLog {
		return this instanceof AuraEventLog;
	}

	isAuraStacksChange(): this is AuraStacksChangeLog {
		return this instanceof AuraStacksChangeLog;
	}

	isMajorCooldownUsed(): this is MajorCooldownUsedLog {
		return this instanceof MajorCooldownUsedLog;
	}

	isCastBegan(): this is CastBeganLog {
		return this instanceof CastBeganLog;
	}

	isCastCompleted(): this is CastCompletedLog {
		return this instanceof CastCompletedLog;
	}

	isCastCancelled(): this is CastCancelledLog {
		return this instanceof CastCancelledLog;
	}

	isAutoDelay(): this is AutoDelayLog {
		return this instanceof AutoDelayLog;
	}

	isStatChange(): this is StatChangeLog {
		return this instanceof StatChangeLog;
	}

	// Group events that happen at the same time.
	static groupDuplicateTimestamps<LogType extends SimLog>(logs: Array<LogType>): Array<Array<LogType>> {
		const grouped: Array<Array<LogType>> = [];
		let curGroup: Array<LogType> = [];

		logs.forEach(log => {
			if (curGroup.length == 0 || log.timestamp == curGroup[0].timestamp) {
				curGroup.push(log);
			} else {
				grouped.push(curGroup);
				curGroup = [log];
			}
		});
		if (curGroup.length > 0) {
			grouped.push(curGroup);
		}

		return grouped;
	}
}

export class DamageDealtLog extends SimLog {
	readonly amount: number;
	readonly type: string;
	readonly miss: boolean;
	readonly hit: boolean;
	readonly crit: boolean;
	readonly crush: boolean;
	readonly glance: boolean;
	readonly dodge: boolean;
	readonly parry: boolean;
	readonly block: boolean;
	readonly tick: boolean;
	readonly partialResist1_4: boolean;
	readonly partialResist2_4: boolean;
	readonly partialResist3_4: boolean;

	constructor(
		params: SimLogParams,
		amount: number,
		type: string,
		miss: boolean,
		crit: boolean,
		crush: boolean,
		glance: boolean,
		dodge: boolean,
		parry: boolean,
		block: boolean,
		tick: boolean,
		partialResist1_4: boolean,
		partialResist2_4: boolean,
		partialResist3_4: boolean,
	) {
		super(params);
		this.amount = amount;
		this.type = type;
		this.miss = miss;
		this.glance = glance;
		this.dodge = dodge;
		this.parry = parry;
		this.block = block;
		this.hit = !miss && !crit;
		this.crit = crit;
		this.crush = crush;
		this.tick = tick;
		this.partialResist1_4 = partialResist1_4;
		this.partialResist2_4 = partialResist2_4;
		this.partialResist3_4 = partialResist3_4;
	}

	isDamage(): boolean {
		return this.type == 'damage';
	}

	isHealing(): boolean {
		return this.type == 'healing';
	}

	isShielding(): boolean {
		return this.type == 'shielding';
	}

	result() {
		const spellSchoolString = typeof this.spellSchool === 'number' ? spellSchoolNames.get(this.spellSchool) : undefined;
		return (
			<>
				{this.isHealing() ? `Healed ` : ''}
				{this.isShielding() ? `Shielded ` : ''}
				{!(this.isHealing() || this.isShielding()) && (
					<>
						{this.miss
							? 'Miss'
							: this.dodge
								? 'Dodge'
								: this.parry
									? 'Parry'
									: this.block
										? this.crit
											? 'Critical Block'
											: this.glance
												? 'Blocked Glance'
												: 'Block'
										: this.glance
											? 'Glance'
											: this.crit
												? 'Crit'
												: this.crush
													? 'Crush'
													: this.tick
														? 'Tick'
														: 'Hit'}
					</>
				)}
				{` `}
				{this.target?.toHTML() || ''}
				{!this.miss && !this.dodge && !this.parry ? (
					<>
						{' '}
						for{' '}
						{this.isHealing() || this.isShielding() ? (
							<strong className={clsx('resource-health')}>{this.amount.toFixed(2)} health</strong>
						) : (
							<strong className={clsx('text-danger', spellSchoolString && `spell-school-${spellSchoolString.toLowerCase()}`)}>
								{this.amount.toFixed(2)} damage
								{spellSchoolString && <> ({spellSchoolString})</>}
							</strong>
						)}
						{this.partialResist1_4 ? (
							<> (25% Resist)</>
						) : this.partialResist2_4 ? (
							<> (50% Resist)</>
						) : this.partialResist3_4 ? (
							<> (75% Resist)</>
						) : (
							''
						)}
						.
					</>
				) : (
					''
				)}
			</>
		);
	}

	toHTML(includeTimestamp = true) {
		const threatPostfix = this.source?.isTarget ? '' : ` (${this.threat.toFixed(2)} Threat)`;
		return (
			<>
				{this.toPrefix(includeTimestamp)} {this.newActionIdLink()} {this.result()}
				{threatPostfix}
			</>
		);
	}

	static build(params: SimLogParams, match: RegExpExecArray): DamageDealtLog {
		const amount = match[16] ? parseFloat(match[16]) : 0;
		const type = match[17] || '';

		return new DamageDealtLog(
			params,
			amount,
			type,
			match[3] == 'Miss',
			match[3] == 'Crit' || match[3] == 'BlockedCrit',
			match[3] == 'Crush',
			match[3] == 'Glance',
			match[3] == 'Dodge',
			match[3] == 'Parry',
			match[3] == 'Block' || match[3] == 'BlockedCrit',
			Boolean(match[2]) && match[2].includes('tick'),
			match[14] == '25',
			match[14] == '50',
			match[14] == '75',
		);
	}
}

export class DpsLog extends SimLog {
	readonly dps: number;

	// Damage events that occurred at the same time as this log.
	readonly damageLogs: Array<DamageDealtLog>;

	constructor(params: SimLogParams, dps: number, damageLogs: Array<DamageDealtLog>) {
		super(params);
		this.dps = dps;
		this.damageLogs = damageLogs;
	}

	static DPS_WINDOW = 15; // Window over which to calculate DPS.
	static fromLogs(damageDealtLogs: Array<DamageDealtLog>): Array<DpsLog> {
		const groupedDamageLogs = SimLog.groupDuplicateTimestamps(damageDealtLogs);

		let curDamageLogs: Array<DamageDealtLog> = [];
		let curDamageTotal = 0;

		return groupedDamageLogs.map(ddLogGroup => {
			ddLogGroup.forEach(ddLog => {
				curDamageLogs.push(ddLog);
				curDamageTotal += ddLog.amount;
			});

			const newStartIdx = curDamageLogs.findIndex(curLog => {
				const inWindow = curLog.timestamp > ddLogGroup[0].timestamp - DpsLog.DPS_WINDOW;
				if (!inWindow) {
					curDamageTotal -= curLog.amount;
				}
				return inWindow;
			});
			if (newStartIdx == -1) {
				curDamageLogs = [];
			} else {
				curDamageLogs = curDamageLogs.slice(newStartIdx);
			}

			const dps = curDamageTotal / DpsLog.DPS_WINDOW;
			if (isNaN(dps)) {
				console.warn('NaN dps!');
			}

			return new DpsLog(
				{
					raw: '',
					logIndex: ddLogGroup[0].logIndex,
					timestamp: ddLogGroup[0].timestamp,
					source: ddLogGroup[0].source,
					target: null,
					actionId: null,
					spellSchool: ddLogGroup[0].spellSchool,
					threat: 0,
				},
				dps,
				ddLogGroup,
			);
		});
	}
}

export class ThreatLogGroup extends SimLog {
	readonly threatBefore: number;
	readonly threatAfter: number;
	readonly logs: Array<SimLog>;

	constructor(params: SimLogParams, threatBefore: number, threatAfter: number, logs: Array<SimLog>) {
		super(params);
		this.threatBefore = threatBefore;
		this.threatAfter = threatAfter;
		this.logs = logs;
	}

	static fromLogs(logs: Array<SimLog>): Array<ThreatLogGroup> {
		const groupedLogs = SimLog.groupDuplicateTimestamps(logs.filter(log => log.threat != 0));
		let curThreat = 0;
		return groupedLogs.map(logGroup => {
			const newThreat = sum(logGroup.map(log => log.threat));
			const threatLog = new ThreatLogGroup(
				{
					raw: '',
					logIndex: logGroup[0].logIndex,
					timestamp: logGroup[0].timestamp,
					source: logGroup[0].source,
					target: logGroup[0].target,
					actionId: null,
					spellSchool: logGroup[0].spellSchool,
					threat: newThreat,
				},
				curThreat,
				curThreat + newThreat,
				logGroup,
			);

			curThreat += newThreat;
			return threatLog;
		});
	}
}

export class AuraEventLog extends SimLog {
	readonly isGained: boolean;
	readonly isFaded: boolean;
	readonly isRefreshed: boolean;

	constructor(params: SimLogParams, isGained: boolean, isFaded: boolean, isRefreshed: boolean) {
		super(params);
		this.isGained = isGained;
		this.isFaded = isFaded;
		this.isRefreshed = isRefreshed;
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)}
				{`  Aura  `}
				{this.isGained ? 'gained' : this.isFaded ? 'faded' : 'refreshed'}: {this.newActionIdLink(true)}.
			</>
		);
	}

	static build(params: SimLogParams, match: RegExpExecArray): AuraEventLog {
		const event = match[1];
		return new AuraEventLog(params, event == 'gained', event == 'faded', event == 'refreshed');
	}
}

export class AuraStacksChangeLog extends SimLog {
	readonly oldStacks: number;
	readonly newStacks: number;

	constructor(params: SimLogParams, oldStacks: number, newStacks: number) {
		super(params);
		this.oldStacks = oldStacks;
		this.newStacks = newStacks;
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} {this.newActionIdLink(true)} stacks: {this.oldStacks} &rarr; {this.newStacks}.
			</>
		);
	}

	static build(params: SimLogParams, match: RegExpExecArray): AuraStacksChangeLog {
		return new AuraStacksChangeLog(params, parseInt(match[2]), parseInt(match[3]));
	}
}

export class AuraUptimeLog extends SimLog {
	readonly gainedAt: number;
	readonly fadedAt: number;
	readonly stacksChange: Array<AuraStacksChangeLog>;

	constructor(params: SimLogParams, fadedAt: number, stacksChange: Array<AuraStacksChangeLog>) {
		super(params);
		this.gainedAt = params.timestamp;
		this.fadedAt = fadedAt;
		this.stacksChange = stacksChange;
	}

	static fromLogs(logs: Array<SimLog>, entity: Entity, encounterDuration: number): Array<AuraUptimeLog> {
		const unmatchedGainedLogs: Array<{ gained: AuraEventLog; stacks: Array<AuraStacksChangeLog> }> = [];
		const uptimeLogs: Array<AuraUptimeLog> = [];

		logs.forEach((log: SimLog) => {
			if (!log.source || !log.source.equals(entity)) {
				return;
			}

			if (log.isAuraStacksChange()) {
				if (log.newStacks <= 0) {
					return;
				}
				const matchingGainedIdx = unmatchedGainedLogs.findIndex(
					gainedLog => gainedLog.gained.actionId!.equals(log.actionId!) || gainedLog.gained.actionId!.equalsIgnoringTag(log.actionId!),
				);
				if (matchingGainedIdx == -1) {
					console.warn('Unmatched aura stacks change log: ' + log.actionId!.name);
					return;
				}
				unmatchedGainedLogs[matchingGainedIdx].stacks.push(log);
				return;
			}

			if (!log.isAuraEvent()) {
				return;
			}

			if (log.isGained) {
				unmatchedGainedLogs.push({ gained: log, stacks: [] });
				return;
			}

			const matchingGainedIdx = unmatchedGainedLogs.findIndex(
				gainedLog => gainedLog.gained.actionId!.equals(log.actionId!) || gainedLog.gained.actionId!.equalsIgnoringTag(log.actionId!),
			);
			if (matchingGainedIdx == -1) {
				console.warn('Unmatched aura faded log: ' + log.actionId!.name);
				return;
			}
			const { gained: gainedLog, stacks: stacksChangeLogs } = unmatchedGainedLogs.splice(matchingGainedIdx, 1)[0];

			uptimeLogs.push(
				new AuraUptimeLog(
					{
						raw: log.raw,
						logIndex: gainedLog.logIndex,
						timestamp: gainedLog.timestamp,
						source: log.source,
						target: log.target,
						actionId: gainedLog.actionId,
						spellSchool: log.spellSchool,
						threat: gainedLog.threat,
					},
					log.timestamp,
					stacksChangeLogs,
				),
			);

			if (log.isRefreshed) {
				unmatchedGainedLogs.push({ gained: log, stacks: [] });
			}
		});

		// Auras active at the end won't have a faded log, so need to add them separately.
		unmatchedGainedLogs.forEach(unmatchedLog => {
			const { gained: gainedLog, stacks: stacksChangeLogs } = unmatchedLog;
			uptimeLogs.push(
				new AuraUptimeLog(
					{
						raw: gainedLog.raw,
						logIndex: gainedLog.logIndex,
						timestamp: gainedLog.timestamp,
						source: gainedLog.source,
						target: gainedLog.target,
						actionId: gainedLog.actionId,
						spellSchool: gainedLog.spellSchool,
						threat: gainedLog.threat,
					},
					encounterDuration,
					stacksChangeLogs,
				),
			);
		});

		uptimeLogs.sort((a, b) => a.gainedAt - b.gainedAt);
		return uptimeLogs;
	}

	// Populates the activeAuras field for all logs using the provided auras.
	static populateActiveAuras(logs: Array<SimLog>, auraLogs: Array<AuraUptimeLog>) {
		// `curAuras` is kept sorted by name as auras come and go, rather than being filtered,
		// copied and re-sorted for every single log. Insertion goes after any equal name so
		// the result matches what a stable sort of the arrival order produced before.
		const curAuras: Array<AuraUptimeLog> = [];
		let auraLogsIndex = 0;
		let changed = true;
		let activeAuras: Array<AuraUptimeLog> = [];

		for (const log of logs) {
			while (auraLogsIndex < auraLogs.length && auraLogs[auraLogsIndex].gainedAt <= log.timestamp) {
				const gained = auraLogs[auraLogsIndex];
				let insertAt = curAuras.length;
				while (insertAt > 0 && stringComparator(curAuras[insertAt - 1].actionId!.name, gained.actionId!.name) > 0) {
					insertAt--;
				}
				curAuras.splice(insertAt, 0, gained);
				auraLogsIndex++;
				changed = true;
			}
			for (let i = curAuras.length - 1; i >= 0; i--) {
				if (curAuras[i].fadedAt <= log.timestamp) {
					curAuras.splice(i, 1);
					changed = true;
				}
			}

			// Consecutive logs almost always see the same set, so they share one array.
			if (changed) {
				activeAuras = curAuras.slice();
				changed = false;
			}
			log.activeAuras = activeAuras;
		}
	}
}

export class ResourceChangedLog extends SimLog {
	readonly resourceType: ResourceType;
	readonly valueBefore: number;
	readonly valueAfter: number;
	readonly isSpend: boolean;
	readonly total: number;

	constructor(params: SimLogParams, resourceType: ResourceType, valueBefore: number, valueAfter: number, isSpend: boolean, total: number) {
		super(params);
		this.resourceType = resourceType;
		this.valueBefore = valueBefore;
		this.valueAfter = valueAfter;
		this.isSpend = isSpend;
		this.total = total;
	}

	toHTML(includeTimestamp = true) {
		const signedDiff = (this.valueAfter - this.valueBefore) * (this.isSpend ? -1 : 1);
		const isHealth = this.resourceType == ResourceType.ResourceTypeHealth;
		const verb = isHealth ? (this.isSpend ? 'Lost' : 'Recovered') : this.isSpend ? 'Spent' : 'Gained';
		const resourceName = resourceNames.get(this.resourceType)!;
		const resourceClass = `resource-${resourceName.replace(/\s/g, '-').toLowerCase()}`;

		return (
			<>
				{this.toPrefix(includeTimestamp)} {verb}{' '}
				<strong className={resourceClass}>
					{signedDiff.toFixed(1)} {resourceName}
				</strong>
				{` from `}
				{this.newActionIdLink()}. ({this.valueBefore.toFixed(1)} &rarr; {this.valueAfter.toFixed(1)})
			</>
		);
	}

	resultString(): string {
		const delta = this.valueAfter - this.valueBefore;
		if (delta < 0) {
			return delta.toFixed(1);
		} else {
			return '+' + delta.toFixed(1);
		}
	}

	static build(params: SimLogParams, match: RegExpExecArray): ResourceChangedLog {
		const [resourceType] = stringToResourceType(match[3]);
		const total = match[8] !== undefined ? parseFloat(match[8]) : 0;
		return new ResourceChangedLog(params, resourceType, parseFloat(match[5]), parseFloat(match[6]), match[1] == 'Spent', total);
	}
}

export class ResourceChangedLogGroup extends SimLog {
	readonly resourceType: ResourceType;
	readonly valueBefore: number;
	readonly valueAfter: number;
	readonly maxValue: number;
	readonly logs: Array<ResourceChangedLog>;

	constructor(params: SimLogParams, resourceType: ResourceType, valueBefore: number, valueAfter: number, maxValue: number, logs: Array<ResourceChangedLog>) {
		super(params);
		this.resourceType = resourceType;
		this.valueBefore = valueBefore;
		this.valueAfter = valueAfter;
		this.maxValue = maxValue;
		this.logs = logs;
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} {resourceNames.get(this.resourceType)}: {this.valueBefore.toFixed(1)} &rarr; {this.valueAfter.toFixed(1)}
			</>
		);
	}

	static fromLogs(logs: Array<SimLog>): Record<ResourceType, Array<ResourceChangedLogGroup>> {
		// Bucket by resource type in one pass. This used to filter the whole resource-log
		// array once per resource type - fifteen scans of the same data - and rebuild the
		// resource type list on every call.
		const byResourceType = new Map<ResourceType, Array<ResourceChangedLog>>();
		for (const log of logs) {
			if (!log.isResourceChanged()) continue;
			const bucket = byResourceType.get(log.resourceType);
			if (bucket) {
				bucket.push(log);
			} else {
				byResourceType.set(log.resourceType, [log]);
			}
		}

		const maxResource = function (logs: ResourceChangedLog[]) {
			let max = 0;
			logs.forEach(l => {
				if (l.total > max) {
					max = l.total;
				}
			});
			return max;
		};
		const results: Partial<Record<ResourceType, Array<ResourceChangedLogGroup>>> = {};
		RESOURCE_TYPES.forEach(resourceType => {
			const groupedLogs = SimLog.groupDuplicateTimestamps(byResourceType.get(resourceType) || []);
			results[resourceType] = groupedLogs.map(
				logGroup =>
					new ResourceChangedLogGroup(
						{
							raw: '',
							logIndex: logGroup[0].logIndex,
							timestamp: logGroup[0].timestamp,
							source: logGroup[0].source,
							target: logGroup[0].target,
							actionId: null,
							spellSchool: logGroup[0].spellSchool,
							threat: 0,
						},
						resourceType,
						logGroup[0].valueBefore,
						logGroup[logGroup.length - 1].valueAfter,
						maxResource(logGroup),
						logGroup,
					),
			);
		});

		return results as Record<ResourceType, Array<ResourceChangedLogGroup>>;
	}
}

export class MajorCooldownUsedLog extends SimLog {
	constructor(params: SimLogParams) {
		super(params);
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} Major cooldown used: {this.newActionIdLink()}.
			</>
		);
	}

	static build(params: SimLogParams): MajorCooldownUsedLog {
		return new MajorCooldownUsedLog(params);
	}
}

export class CastBeganLog extends SimLog {
	readonly manaCost: number;
	readonly castTime: number;
	readonly gcd: number;
	readonly effectiveTime: number;

	constructor(params: SimLogParams, manaCost: number, castTime: number, gcd: number, effectiveTime: number) {
		super(params);
		this.manaCost = manaCost;
		this.castTime = castTime;
		this.gcd = gcd;
		this.effectiveTime = effectiveTime;
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} Casting {this.newActionIdLink()} (Cast time: {this.castTime.toFixed(2)}s, Cost: {this.manaCost.toFixed(1)}{' '}
				Mana).
			</>
		);
	}

	static build(params: SimLogParams, match: RegExpExecArray): CastBeganLog {
		let castTime = parseFloat(match[3]);
		if (match[4] == 'ms') {
			castTime /= 1000;
		}
		let gcd = parseFloat(match[5]);
		if (match[6] == 'ms') {
			gcd /= 1000;
		}
		let effectiveTime = parseFloat(match[7]);
		if (match[8] == 'ms') {
			effectiveTime /= 1000;
		}
		return new CastBeganLog(params, parseFloat(match[2]), castTime, gcd, effectiveTime);
	}
}

export class CastCancelledLog extends SimLog {
	readonly cancelTime: number;

	constructor(params: SimLogParams, cancelTime: number) {
		super(params);
		this.cancelTime = cancelTime;
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} Cancelled {this.newActionIdLink()} after {this.cancelTime.toFixed(2)}s.
			</>
		);
	}

	static build(params: SimLogParams, match: RegExpExecArray): CastCancelledLog {
		let castProgress = parseFloat(match[2]);
		if (match[3] == 'ms') {
			castProgress /= 1000;
		}
		return new CastCancelledLog(params, castProgress);
	}
}

export class CastCompletedLog extends SimLog {
	constructor(params: SimLogParams) {
		super(params);
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} Completed cast {this.actionId!.name}.
			</>
		);
	}

	static build(params: SimLogParams): CastCompletedLog {
		return new CastCompletedLog(params);
	}
}

// AutoDelayLog records the gap between when an auto-attack would
// naturally have fired and when it actually did, emitted from the auto's
// ApplyEffects when the delay exceeds 1ms. Attached to the matching
// CastLog at construction time.
export class AutoDelayLog extends SimLog {
	readonly delay: number; // seconds, used for block width math
	readonly delayText: string; // pre-formatted for display: rounded ms below 1s, 2-decimal s at 1s+
	readonly readyAtLogText: string; // pre-formatted for display in log view
	readonly readyAtTooltip: string; // pre-formatted for display in tooltip

	constructor(params: SimLogParams, delay: number, delayText: string, readyAtLogText: string, readyAtTooltip: string) {
		super(params);
		this.delay = delay;
		this.delayText = delayText;
		this.readyAtLogText = readyAtLogText;
		this.readyAtTooltip = readyAtTooltip;
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} {this.newActionIdLink()} delayed by {this.delayText}, was ready at {this.readyAtLogText}.
			</>
		);
	}

	static build(params: SimLogParams, match: RegExpExecArray): AutoDelayLog {
		let delay = parseFloat(match[2]);
		if (match[3] == 'ms') {
			delay /= 1000;
		}
		const delayText = delay >= 1 ? `${delay.toFixed(2)}s` : `${Math.round(delay * 1000)}ms`;

		const readyAtMinute = match[4];
		const readyAtSeconds = parseFloat(match[5]);
		let readyAtTotal = readyAtSeconds;
		if (readyAtMinute && readyAtMinute.endsWith('m')) {
			readyAtTotal = parseFloat(readyAtMinute.slice(0, -1)) * 60 + readyAtSeconds;
		}
		const readyAtLogText = readyAtTotal >= 1 ? `${readyAtMinute}${readyAtSeconds.toFixed(2)}s` : `${Math.round(readyAtTotal * 1000)}ms`;
		const readyAtTooltip = readyAtTotal >= 1 ? `${readyAtTotal.toFixed(2)}s` : `${Math.round(readyAtTotal * 1000)}ms`;

		return new AutoDelayLog(params, delay, delayText, readyAtLogText, readyAtTooltip);
	}
}

export class CastLog extends SimLog {
	readonly castTime: number;
	readonly gcd: number;
	readonly effectiveTime: number;
	readonly travelTime: number;
	readonly cancelTime: number;
	readonly delay: number;
	readonly delayText: string;
	readonly readyAtText: string;

	readonly castBeganLog: CastBeganLog;
	readonly castCancelledLog: CastCancelledLog | null;
	readonly castCompletedLog: CastCompletedLog | null;

	// All instances of damage dealt from the completion of this cast until the completion of the next cast.
	readonly damageDealtLogs: Array<DamageDealtLog>;

	constructor(
		castBeganLog: CastBeganLog,
		castCompletedLog: CastCompletedLog | null,
		castCancelledLog: CastCancelledLog | null,
		autoDelayLog: AutoDelayLog | null,
		damageDealtLogs: Array<DamageDealtLog>,
	) {
		super({
			raw: castBeganLog.raw,
			logIndex: castBeganLog.logIndex,
			timestamp: castBeganLog.timestamp,
			source: castBeganLog.source,
			target: castBeganLog.target,
			actionId: castCompletedLog?.actionId || castCancelledLog?.actionId || castBeganLog.actionId, // Use completed log because of arcane blast
			spellSchool: castCompletedLog?.spellSchool || castCancelledLog?.spellSchool || castBeganLog.spellSchool,
			threat: castCompletedLog?.threat || castCancelledLog?.threat || castBeganLog.threat,
		});
		this.castTime = castBeganLog.castTime;
		this.gcd = castBeganLog.gcd;
		this.effectiveTime = castBeganLog.effectiveTime;
		this.cancelTime = castCancelledLog?.cancelTime || 0;
		this.delay = autoDelayLog?.delay ?? 0;
		this.delayText = autoDelayLog?.delayText ?? '';
		this.readyAtText = autoDelayLog?.readyAtTooltip ?? '';

		this.castBeganLog = castBeganLog;
		this.castCompletedLog = castCompletedLog;
		this.castCancelledLog = castCancelledLog;
		this.damageDealtLogs = damageDealtLogs;

		if (this.castCompletedLog && this.castBeganLog) {
			this.castTime = this.castCompletedLog.timestamp - this.castBeganLog.timestamp;
		}
		if (this.castCancelledLog) {
			this.cancelTime = this.castCancelledLog.cancelTime;
		}
		if (
			this.castCompletedLog &&
			this.damageDealtLogs.length >= 1 &&
			this.castCompletedLog.timestamp < this.damageDealtLogs[0].timestamp &&
			!this.damageDealtLogs[0].tick
		) {
			this.travelTime = this.damageDealtLogs[0].timestamp - this.castCompletedLog.timestamp;
		} else {
			this.travelTime = 0;
		}
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} Casting {this.actionId!.name} (Cast time = {this.castTime.toFixed(2)}s).
			</>
		);
	}

	totalDamage(): number {
		return sum(this.damageDealtLogs.map(ddl => ddl.amount));
	}

	static fromLogs(logs: Array<SimLog>): Array<CastLog> {
		// One classification pass instead of five full scans of the same array.
		const castBeganLogs: Array<CastBeganLog> = [];
		const castCompletedLogs: Array<CastCompletedLog> = [];
		const castCancelledLogs: Array<CastCancelledLog> = [];
		const damageDealtLogs: Array<DamageDealtLog> = [];
		const autoDelayLogs: Array<AutoDelayLog> = [];
		for (const log of logs) {
			// instanceof rather than the isX() guards, which are the same test but declared as
			// `this is X`; chaining those narrows negatively and leaves the compiler holding
			// `never` by the third branch.
			if (log instanceof CastBeganLog) castBeganLogs.push(log);
			else if (log instanceof CastCompletedLog) castCompletedLogs.push(log);
			else if (log instanceof CastCancelledLog) castCancelledLogs.push(log);
			else if (log instanceof DamageDealtLog) damageDealtLogs.push(log);
			else if (log instanceof AutoDelayLog) autoDelayLogs.push(log);
		}

		const toBucketKey = (actionId: ActionId) => {
			if (actionId.spellId == 30451 || actionId.spellId == 127632) {
				// Arcane Blast is unique because it can finish its cast as a different
				// spell than it started (if stacks drop).
				// Also handle Shadow's Cascade for bouncing
				return actionId.toStringIgnoringTag();
			} else if (actionId.spellId === 27014) {
				// Raptor Strike should be grouped with regular melee swings
				return 'other-3-1';
			} else {
				return actionId.toString();
			}
		};
		const castBeganLogsByAbility = bucket(castBeganLogs, log => toBucketKey(log.actionId!));
		const castCompletedLogsByAbility = bucket(castCompletedLogs, log => toBucketKey(log.actionId!));
		const castCancelledLogsByAbility = bucket(castCancelledLogs, log => toBucketKey(log.actionId!));
		const damageDealtLogsByAbility = bucket(damageDealtLogs, log => toBucketKey(log.actionId!));
		const autoDelayLogsByAbility = bucket(autoDelayLogs, log => toBucketKey(log.actionId!));

		const castLogs: Array<CastLog> = [];
		Object.keys(castBeganLogsByAbility).forEach(bucketKey => {
			const abilityCastsBegan = castBeganLogsByAbility[bucketKey]!;
			const abilityCastsCompleted = castCompletedLogsByAbility[bucketKey];
			const abilityCastsCancelled = castCancelledLogsByAbility[bucketKey];
			const abilityDamageDealt = damageDealtLogsByAbility[bucketKey];
			const abilityAutoDelay = autoDelayLogsByAbility[bucketKey];

			let ddIdx = 0;
			let adIdx = 0;
			let castSkipIdx = 0;
			for (let cbIdx = 0; cbIdx < abilityCastsBegan.length; cbIdx++) {
				const cbLog = abilityCastsBegan[cbIdx];

				// Assume cast completed log is the same index because they always come in pairs.
				// Only exception is final pair, where there might be a cast began without a cast completed.
				let ccLog: CastCompletedLog | null = null;
				let cCancelLog: CastCancelledLog | null = null;
				let nextCcLog: CastCompletedLog | null = null;
				if (abilityCastsCompleted && cbIdx < abilityCastsCompleted.length) {
					ccLog = abilityCastsCompleted[cbIdx + castSkipIdx];
					if (cbIdx + castSkipIdx + 1 < abilityCastsCompleted.length) {
						nextCcLog = abilityCastsCompleted[cbIdx + castSkipIdx + 1];
					}
				}

				if (abilityCastsCancelled) {
					const cancelledCast =
						abilityCastsCancelled.find(
							cancelLog =>
								cancelLog.timestamp >= cbLog.timestamp &&
								(!abilityCastsBegan[cbIdx + 1] || cancelLog.timestamp <= abilityCastsBegan[cbIdx + 1].timestamp),
						) || null;
					if (cancelledCast) {
						cCancelLog = cancelledCast;
						castSkipIdx--;
					}
				}
				if (cbLog.actionId?.spellId === 2912) {
				}

				// Find all damage dealt logs between the cur and next cast completed logs.
				const ddLogs = [];
				while (
					!cCancelLog &&
					abilityDamageDealt &&
					ddIdx < abilityDamageDealt.length &&
					(!nextCcLog || abilityDamageDealt[ddIdx].timestamp < nextCcLog.timestamp)
				) {
					ddLogs.push(abilityDamageDealt[ddIdx]);
					ddIdx++;
				}

				// At most one AutoDelay log per cast; pick the next one inside this cast's window.
				const nextCbLog = abilityCastsBegan[cbIdx + 1];
				let adLog: AutoDelayLog | null = null;
				if (
					abilityAutoDelay &&
					adIdx < abilityAutoDelay.length &&
					abilityAutoDelay[adIdx].timestamp >= cbLog.timestamp &&
					(!nextCbLog || abilityAutoDelay[adIdx].timestamp < nextCbLog.timestamp)
				) {
					adLog = abilityAutoDelay[adIdx];
					adIdx++;
				}

				castLogs.push(new CastLog(cbLog, ccLog, cCancelLog, adLog, ddLogs));
			}
		});

		castLogs.sort((a, b) => a.timestamp - b.timestamp);
		return castLogs;
	}
}

export class StatChangeLog extends SimLog {
	readonly isGain: boolean;
	readonly stats: string;

	constructor(params: SimLogParams, isGain: boolean, stats: string) {
		super(params);
		this.isGain = isGain;
		this.stats = stats;
	}

	toHTML(includeTimestamp = true) {
		if (this.isGain) {
			return (
				<>
					{this.toPrefix(includeTimestamp)} Gained {this.stats} from {this.newActionIdLink()}.
				</>
			);
		} else {
			return (
				<>
					{this.toPrefix(includeTimestamp)} Lost {this.stats} from fading {this.newActionIdLink()}.
				</>
			);
		}
	}

	static build(params: SimLogParams, match: RegExpExecArray): StatChangeLog {
		const sign = match[1] == 'Lost' ? -1 : 1;
		return new StatChangeLog(params, sign == 1, match[4]);
	}
}

// Preamble patterns, hoisted so parseAll does not allocate a fresh RegExp per line.
const RESOURCE_TYPES = (getEnumValues(ResourceType) as Array<ResourceType>).filter(val => val != ResourceType.ResourceTypeNone);

const TIMESTAMP_PREFIX_REGEX = /(\[[0-9.-]+\]) (\[[0-9a-zA-Z\s\-()#]+\])?(.*)/;

const SPELL_SCHOOL_REGEX = / \(SpellSchool: (-?[0-9]+)\)/;
const THREAT_REGEX = / \(Threat: (-?[0-9]+\.[0-9]+)\)/;
const TIMESTAMP_REGEX = /\[(-?[0-9]+\.[0-9]+)\]\w*(.*)/;

type LogMatcher = {
	// Literals the pattern requires: a line must contain at least one of them to be worth
	// testing. Skipping a matcher whose guard fails can therefore never skip a match the
	// pattern would have made, which is what keeps this dispatch equivalent to running every
	// pattern in order - it just stops paying for the ones that cannot match. The damage
	// pattern is both the first tried and the most expensive, so guarding it is most of the
	// win. assertGuardsAreNecessary() below checks each literal really is in its pattern.
	guard: Array<string>;
	regex: RegExp;
	// Some patterns match lines they cannot actually build from; those fall through to the
	// next matcher, exactly as the old `parse` chain did by returning null.
	valid?: (match: RegExpExecArray) => boolean;
	// The fragment naming this line's action, e.g. '{SpellID: 48707}'.
	idString: (match: RegExpExecArray) => string;
	build: (params: SimLogParams, match: RegExpExecArray) => SimLog;
};

// Order is load-bearing: first match wins, and it runs most to least common.
const LOG_MATCHERS: Array<LogMatcher> = [
	{
		guard: ['Miss', 'Hit', 'Crit', 'Crush', 'Glance', 'Dodge', 'Parry', 'Block'],
		regex: /] (.*?) (tick )?((Miss)|(Hit)|(BlockedCrit)|(Crit)|(Crush)|(Glance)|(Dodge)|(Parry)|(Block))( \((\d+)% Resist\))?( for (\d+\.\d+) ((damage)|(healing)|(shielding)))?/,
		idString: match => match[1],
		build: (params, match) => DamageDealtLog.build(params, match),
	},
	{
		guard: [' from '],
		regex: /(Gained|Spent) (\d+\.?\d*) (\S.+?\S) from (.*?) \((\d+\.?\d*) --> (\d+\.?\d*)\)( of (\d+\.?\d*) total)?/,
		idString: match => match[4],
		build: (params, match) => ResourceChangedLog.build(params, match),
	},
	{
		guard: ['Aura '],
		regex: /Aura ((gained)|(faded)|(refreshed)): (.*)/,
		valid: match => Boolean(match[5]),
		idString: match => match[5],
		build: (params, match) => AuraEventLog.build(params, match),
	},
	{
		guard: [' stacks: '],
		regex: /(.*) stacks: ([0-9]+) --> ([0-9]+)/,
		valid: match => Boolean(match[1]),
		idString: match => match[1],
		build: (params, match) => AuraStacksChangeLog.build(params, match),
	},
	{
		guard: ['Major cooldown used: '],
		regex: /Major cooldown used: (.*)/,
		idString: match => match[1],
		build: params => MajorCooldownUsedLog.build(params),
	},
	{
		guard: ['Casting '],
		regex: /Casting (.*) \(Cost = (\d+\.?\d*), Cast Time = (\d+\.?\d*)(m?s), GCD = (\d+\.?\d*)(m?s), Effective Time = (\d+\.?\d*)(m?s)\)/,
		idString: match => match[1],
		build: (params, match) => CastBeganLog.build(params, match),
	},
	{
		guard: ['Completed cast '],
		regex: /Completed cast (.*)/,
		idString: match => match[1],
		build: params => CastCompletedLog.build(params),
	},
	{
		guard: [' delayed by '],
		regex: /] (.*?) delayed by (\d+\.?\d*)(m?s), was ready at (\d*?m?)(\d+\.?\d*)(m?s)/,
		idString: match => match[1],
		build: (params, match) => AutoDelayLog.build(params, match),
	},
	{
		guard: [' from '],
		regex: /((Gained)|(Lost)) ({.*}) from (fading )?(.*)/,
		idString: match => match[6],
		build: (params, match) => StatChangeLog.build(params, match),
	},
	{
		guard: ['Queueing up '],
		regex: /Queueing up (.*?) to cast at (\d*?m?)(\d+\.?\d*)(m?s)\./,
		idString: match => match[1],
		build: (params, match) => SpellQueuedLog.build(params, match),
	},
	{
		guard: [' failed to cast: '],
		regex: /] (.*?) failed to cast: (.*)/,
		idString: match => match[1],
		build: (params, match) => CastFailedLog.build(params, match),
	},
	{
		guard: [' pushed back '],
		regex: /] (.*?) pushed back (\d+\.?\d*)(m?s) while ((casting)|(channeling))/,
		idString: match => match[1],
		build: (params, match) => CastPushbackLog.build(params, match),
	},
];

type MatchedLine = { matcher: LogMatcher; match: RegExpExecArray };
type PendingLog = { params: SimLogParams; matcher: LogMatcher | null; match: RegExpExecArray | null; key: string };
type ActionIdRequest = { logString: string; playerIndex: number | undefined };

// The dispatch is only equivalent to trying every pattern in order if each guard literal is
// something its own pattern requires. That is a property of the table, so check it here
// rather than trusting the comment. Necessary, not sufficient: a literal inside an optional
// group would pass this and still not be required, so guards stay hand-reviewed too.
function assertGuardsAreNecessary() {
	for (const matcher of LOG_MATCHERS) {
		for (const literal of matcher.guard) {
			if (!matcher.regex.source.includes(literal)) {
				throw new Error(`Log matcher guard '${literal}' is absent from its pattern ${matcher.regex.source}`);
			}
		}
	}
}

function matchLogLine(raw: string): MatchedLine | null {
	for (const matcher of LOG_MATCHERS) {
		let possible = false;
		for (const literal of matcher.guard) {
			if (raw.includes(literal)) {
				possible = true;
				break;
			}
		}
		if (!possible) continue;
		const match = matcher.regex.exec(raw);
		if (match && (!matcher.valid || matcher.valid(match))) {
			return { matcher, match };
		}
	}
	return null;
}

// Records when a spell is queued to fire at a future time, typically right
// before a hardcast or auto-attack completes.
export class SpellQueuedLog extends SimLog {
	readonly fireAt: number; // seconds
	readonly fireAtText: string; // pre-formatted for display

	constructor(params: SimLogParams, fireAt: number, fireAtText: string) {
		super(params);
		this.fireAt = fireAt;
		this.fireAtText = fireAtText;
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} Queueing {this.newActionIdLink()} to cast at {this.fireAtText}.
			</>
		);
	}

	static build(params: SimLogParams, match: RegExpExecArray): SpellQueuedLog {
		const minutePart = match[2];
		const seconds = parseFloat(match[3]);
		let fireAt = seconds;
		if (minutePart && minutePart.endsWith('m')) {
			fireAt = parseFloat(minutePart.slice(0, -1)) * 60 + seconds;
		} else if (match[4] == 'ms') {
			fireAt = seconds / 1000;
		}
		const fireAtText = fireAt >= 1 ? `${minutePart}${seconds.toFixed(2)}s` : `${Math.round(fireAt * 1000)}ms`;

		return new SpellQueuedLog(params, fireAt, fireAtText);
	}
}

// Go's time.Duration.String() can emit nanosecond-precision strings like
// "29.999999999s" or "33.217000001s". Round these to something readable
// without disturbing already-clean values like "30s" or "1.5s".
const reformatGoDurations = (text: string): string =>
	text.replace(/(\d+m)?(\d+\.\d{3,})(m?s)/g, (_match, minute: string | undefined, value: string, unit: string) => {
		const seconds = parseFloat(value);
		if (unit === 'ms') {
			return `${minute ?? ''}${Math.round(seconds)}ms`;
		}
		return `${minute ?? ''}${seconds.toFixed(2)}s`;
	});

// The sim's cast-failure reasons all append ", curTime = <duration>", which
// is redundant with the log line's own timestamp prefix.
const stripCurTime = (text: string): string => text.replace(/, curTime = [\dms.h]+$/, '');

// Records when a cast attempt was rejected (cooldown, GCD, already casting,
// resource cost, etc.). The reason text comes from the sim log, with the
// redundant ", curTime = ..." stripped, durations rounded, and embedded
// action IDs resolved to names.
export class CastFailedLog extends SimLog {
	reason: string;

	constructor(params: SimLogParams, reason: string) {
		super(params);
		this.reason = reason;
	}

	toHTML(includeTimestamp = true) {
		return (
			<>
				{this.toPrefix(includeTimestamp)} Failed to cast {this.newActionIdLink()}: {this.reason}.
			</>
		);
	}

	// The reason can itself name spells, which parseAll resolves in a second pass; this one is
	// synchronous.
	static build(params: SimLogParams, match: RegExpExecArray): CastFailedLog {
		return new CastFailedLog(params, reformatGoDurations(stripCurTime(match[2])));
	}
}

// Records pushback events emitted when the player takes damage while
// hardcasting or channeling, delaying the cast or shortening the channel.
export class CastPushbackLog extends SimLog {
	readonly pushback: number; // seconds
	readonly pushbackText: string; // pre-formatted for display
	readonly isChanneling: boolean;

	constructor(params: SimLogParams, pushback: number, pushbackText: string, isChanneling: boolean) {
		super(params);
		this.pushback = pushback;
		this.pushbackText = pushbackText;
		this.isChanneling = isChanneling;
	}

	toHTML(includeTimestamp = true) {
		return this.isChanneling ? (
			<>
				{this.toPrefix(includeTimestamp)} {this.newActionIdLink()} lost {this.pushbackText} while channeling due to pushback.
			</>
		) : (
			<>
				{this.toPrefix(includeTimestamp)} {this.newActionIdLink()} pushed back {this.pushbackText} while casting.
			</>
		);
	}

	static build(params: SimLogParams, match: RegExpExecArray): CastPushbackLog {
		let pushback = parseFloat(match[2]);
		if (match[3] == 'ms') {
			pushback /= 1000;
		}
		const pushbackText = pushback >= 1 ? `${pushback.toFixed(2)}s` : `${Math.round(pushback * 1000)}ms`;

		return new CastPushbackLog(params, pushback, pushbackText, match[4] == 'channeling');
	}
}
