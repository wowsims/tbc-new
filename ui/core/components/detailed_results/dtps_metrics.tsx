import { ActionMetrics } from '../../proto_utils/sim_result';
import i18n from '../../../i18n/config';
import { formatToCompactNumber, formatToNumber, formatToPercent } from '../../utils';
import { MetricsCombinedTooltipTable } from './metrics_table/metrics_combined_tooltip_table';
import { ColumnSortType, MetricsTable } from './metrics_table/metrics_table';
import { MetricsTotalBar } from './metrics_table/metrics_total_bar';
import { ResultComponentConfig, SimResultData } from './result_component';

export class DtpsMetricsTable extends MetricsTable<ActionMetrics> {
	maxDtpsAmount: number | null = null;
	constructor(config: ResultComponentConfig) {
		config.rootCssClass = 'dtps-metrics-root';
		config.resultsEmitter.on((_, resultData) => {
			const lastResult = resultData
				? this.getGroupedMetrics(resultData)
						.filter(g => g.length)
						.map(groups => this.mergeMetrics(groups))
				: undefined;
			this.maxDtpsAmount = Math.max(...(lastResult || []).map(a => a.damage));
		});
		super(config, [
			MetricsTable.nameCellConfig((metric: ActionMetrics) => {
				return {
					name: metric.name,
					actionId: metric.actionId,
					metricType: metric.constructor?.name,
				};
			}),
			{
				name: i18n.t('results_tab.details.columns.damage_taken'),
				headerCellClass: 'text-center metrics-table-cell--primary-metric',
				columnClass: 'metrics-table-cell--primary-metric',
				getValue: (metric: ActionMetrics) => metric.avgDamage,
				fillCell: (metric: ActionMetrics, cellElem: HTMLElement) => {
					cellElem.appendChild(
						<MetricsTotalBar
							spellSchool={metric.spellSchool}
							percentage={metric.totalDamageTakenPercent}
							max={this.maxDtpsAmount}
							total={metric.avgDamage}
							value={metric.damage}
						/>,
					);

					const hitValues = metric.damageDone.hit;
					const resistedHitValues = metric.damageDone.resistedHit;
					const critHitValues = metric.damageDone.critHit;
					const resistedCritHitValues = metric.damageDone.resistedCritHit;
					const tickValues = metric.damageDone.tick;
					const resistedTickValues = metric.damageDone.resistedTick;
					const critTickValues = metric.damageDone.critTick;
					const resistedCritTickValues = metric.damageDone.resistedCritTick;
					const glanceValues = metric.damageDone.glance;
					const blockValues = metric.damageDone.block;
					const blockedCritValues = metric.damageDone.blockedCrit;

					cellElem.appendChild(
						<MetricsCombinedTooltipTable
							tooltipElement={cellElem}
							headerValues={[, i18n.t('results_tab.details.tooltip_table.amount')]}
							groups={[
								{
									spellSchool: metric.spellSchool,
									total: metric.damage,
									totalPercentage: 100,
									data: [
										{
											name: i18n.t('results_tab.details.attack_types.hit'),
											...hitValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.resisted_hit'),
											...resistedHitValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.critical_hit'),
											...critHitValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.resisted_critical_hit'),
											...resistedCritHitValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.tick'),
											...tickValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.resisted_tick'),
											...resistedTickValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.critical_tick'),
											...critTickValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.resisted_critical_tick'),
											...resistedCritTickValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.glancing_blow'),
											...glanceValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.blocked_hit'),
											...blockValues,
										},
										{
											name: i18n.t('results_tab.details.attack_types.blocked_critical_hit'),
											...blockedCritValues,
										},
									],
								},
							]}
						/>,
					);
				},
			},
			{
				name: i18n.t('results_tab.details.columns.casts'),
				getValue: (metric: ActionMetrics) => metric.casts,
				fillCell: (metric: ActionMetrics, cellElem: HTMLElement) => {
					cellElem.appendChild(<>{formatToNumber(metric.casts, { fallbackString: '-' })}</>);

					if ((!metric.landedHits && !metric.totalMisses) || metric.isPassiveAction) return;
					const relativeHitPercent = ((metric.landedHits || metric.casts) / ((metric.landedHits || metric.casts) + metric.totalMisses)) * 100;
					cellElem.appendChild(
						<MetricsCombinedTooltipTable
							tooltipElement={cellElem}
							groups={[
								{
									spellSchool: metric.spellSchool,
									total: metric.casts,
									totalPercentage: 100,
									data: [
										{
											name: i18n.t('results_tab.details.attack_types.hit') + 's',
											value: metric.landedHits || metric.casts - metric.totalMisses,
											percentage: relativeHitPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.miss'),
											value: metric.misses,
											percentage: metric.missPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.parry'),
											value: metric.parries,
											percentage: metric.parryPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.dodge'),
											value: metric.dodges,
											percentage: metric.dodgePercent,
										},
									],
								},
							]}
						/>,
					);
				},
			},
			{
				name: i18n.t('results_tab.details.columns.avg_cast'),
				tooltip: i18n.t('results_tab.details.tooltips.damage_avg_cast_tooltip'),
				getValue: (metric: ActionMetrics) => {
					if (metric.isPassiveAction) return 0;
					return metric.avgCastHit || metric.avgCastTick;
				},
				fillCell: (metric: ActionMetrics, cellElem: HTMLElement) => {
					cellElem.appendChild(
						<>
							{formatToCompactNumber(metric.avgCastHit || metric.avgCastTick, { fallbackString: '-' })}
							{metric.avgCastHit && metric.avgCastTick ? <> ({formatToCompactNumber(metric.avgCastTick, { fallbackString: '-' })})</> : undefined}
						</>,
					);
				},
			},
			{
				name: i18n.t('results_tab.details.columns.hits'),
				getValue: (metric: ActionMetrics) => metric.landedHits || metric.landedTicks,
				fillCell: (metric: ActionMetrics, cellElem: HTMLElement) => {
					cellElem.appendChild(
						<>
							{formatToNumber(metric.landedHits || metric.landedTicks, { fallbackString: '-' })}
							{metric.landedHits && metric.landedTicks ? <> ({formatToNumber(metric.landedTicks, { fallbackString: '-' })})</> : undefined}
						</>,
					);
					if (!metric.landedHits && !metric.landedTicks) return;

					const relativeHitPercent = ((metric.hits - metric.resistedHits) / metric.landedHits) * 100;
					const relativeResistedHitPercent = (metric.resistedHits / metric.landedHits) * 100;
					const relativeCritPercent = ((metric.crits - metric.resistedCrits) / metric.landedHits) * 100;
					const relativeResistedCritPercent = (metric.resistedCrits / metric.landedHits) * 100;
					const relativeTickPercent = ((metric.ticks - metric.resistedTicks) / metric.landedTicks) * 100;
					const relativeResistedTickPercent = (metric.resistedTicks / metric.landedTicks) * 100;
					const relativeCritTickPercent = ((metric.critTicks - metric.resistedCritTicks) / metric.landedTicks) * 100;
					const relativeResistedCritTickPercent = (metric.resistedCritTicks / metric.landedTicks) * 100;
					const relativeGlancePercent = (metric.glances / metric.landedHits) * 100;
					const relativeBlockPercent = (metric.blocks / metric.landedHits) * 100;
					const relativeBlockedCritPercent = (metric.blockedCrits / metric.landedHits) * 100;

					cellElem.appendChild(
						<MetricsCombinedTooltipTable
							tooltipElement={cellElem}
							groups={[
								{
									spellSchool: metric.spellSchool,
									total: metric.landedHits,
									totalPercentage: 100,
									data: [
										{
											name: i18n.t('results_tab.details.attack_types.hit'),
											value: metric.hits - metric.resistedHits,
											percentage: relativeHitPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.resisted_hit'),
											value: metric.resistedHits,
											percentage: relativeResistedHitPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.critical_hit'),
											value: metric.crits - metric.resistedCrits,
											percentage: relativeCritPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.blocked_critical_hit'),
											value: metric.blockedCrits,
											percentage: relativeBlockedCritPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.resisted_critical_hit'),
											value: metric.resistedCrits,
											percentage: relativeResistedCritPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.glancing_blow'),
											value: metric.glances,
											percentage: relativeGlancePercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.blocked_hit'),
											value: metric.blocks,
											percentage: relativeBlockPercent,
										},
									],
								},
								{
									spellSchool: metric.spellSchool,
									total: metric.landedTicks,
									totalPercentage: 100,
									data: [
										{
											name: i18n.t('results_tab.details.attack_types.tick'),
											value: metric.ticks - metric.resistedTicks,
											percentage: relativeTickPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.resisted_tick'),
											value: metric.resistedTicks,
											percentage: relativeResistedTickPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.critical_tick'),
											value: metric.critTicks - metric.resistedCritTicks,
											percentage: relativeCritTickPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.resisted_critical_tick'),
											value: metric.resistedCritTicks,
											percentage: relativeResistedCritTickPercent,
										},
									],
								},
							]}
						/>,
					);
				},
			},
			{
				name: i18n.t('results_tab.details.columns.avg_hit'),
				getValue: (metric: ActionMetrics) => metric.avgHit || metric.avgTick,
				fillCell: (metric: ActionMetrics, cellElem: HTMLElement) => {
					cellElem.appendChild(
						<>
							{formatToCompactNumber(metric.avgHit || metric.avgTick, { fallbackString: '-' })}
							{metric.avgHit && metric.avgTick ? <> ({formatToCompactNumber(metric.avgTick, { fallbackString: '-' })})</> : undefined}
						</>,
					);
				},
			},
			{
				name: i18n.t('results_tab.details.columns.miss_percent'),
				tooltip: i18n.t('results_tab.details.tooltips.hit_miss_percent_tooltip'),
				getValue: (metric: ActionMetrics) => metric.totalMissesPercent,
				fillCell: (metric: ActionMetrics, cellElem: HTMLElement) => {
					cellElem.appendChild(<>{formatToPercent(metric.totalMissesPercent, { fallbackString: '-' })}</>);
					if (!metric.totalMissesPercent) return;

					cellElem.appendChild(
						<MetricsCombinedTooltipTable
							tooltipElement={cellElem}
							groups={[
								{
									spellSchool: metric.spellSchool,
									total: metric.totalMisses,
									totalPercentage: metric.totalMissesPercent,
									data: [
										{
											name: i18n.t('results_tab.details.attack_types.miss'),
											value: metric.misses,
											percentage: metric.missPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.parry'),
											value: metric.parries,
											percentage: metric.parryPercent,
										},
										{
											name: i18n.t('results_tab.details.attack_types.dodge'),
											value: metric.dodges,
											percentage: metric.dodgePercent,
										},
									],
								},
							]}
						/>,
					);
				},
			},
			{
				name: i18n.t('results_tab.details.columns.crit_percent'),
				getValue: (metric: ActionMetrics) => metric.critPercent + metric.blockedCritPercent || metric.critTickPercent,
				getDisplayString: (metric: ActionMetrics) =>
					`${formatToPercent(metric.critPercent + metric.blockedCritPercent || metric.critTickPercent, { fallbackString: '-' })}${
						metric.critPercent + metric.blockedCritPercent && metric.critTickPercent
							? ` (${formatToPercent(metric.critTickPercent, { fallbackString: '-' })})`
							: ''
					}`,
			},
			{
				name: i18n.t('results_tab.details.columns.dtps'),
				sort: ColumnSortType.Descending,
				headerCellClass: 'text-body',
				columnClass: 'text-success',
				getValue: (metric: ActionMetrics) => metric.dps,
				getDisplayString: (metric: ActionMetrics) => formatToNumber(metric.dps, { minimumFractionDigits: 2, fallbackString: '-' }),
			},
		]);
	}

	getGroupedMetrics(resultData: SimResultData): Array<Array<ActionMetrics>> {
		const players = resultData.result.getRaidIndexedPlayers(resultData.filter);
		if (players.length != 1) {
			return [];
		}
		const player = players[0];

		const targets = resultData.result.getTargets(resultData.filter);
		const targetActions = targets.map(target => target.getDamageActions().map(action => action.forTarget({ player: player.unitIndex }))).flat();
		const actionGroups = ActionMetrics.groupById(targetActions);

		return actionGroups;
	}

	mergeMetrics(metrics: Array<ActionMetrics>): ActionMetrics {
		// TODO: Use NPC ID here instead of pet ID.
		return ActionMetrics.merge(metrics, {
			removeTag: true,
			actionIdOverride: metrics[0].unit?.petActionId || undefined,
		});
	}
}
