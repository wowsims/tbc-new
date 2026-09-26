import { BulkStatConstraint, BulkStatConstraintOp } from '@generated/proto/api';
import { PseudoStat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { newStatConstraint, STAT_CONSTRAINT_OP_SYMBOLS } from '@sim/bulk/stat_constraints';
import { usePlayer } from '@sim/context/SimHostContext';
import { Button } from '@ui-kit/Button';
import { Input, Select } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';

import { useBulkState } from '../../hooks/useBulkState';
import { setBulkStatConstraints } from '../../model/settings';
import {
	constraintUnitStat,
	SELECTABLE_STATS,
	STAT_CONSTRAINT_OPS,
	unitStatFromOptionValue,
	unitStatOptionValue,
	unitStatShortLabel,
	withConstraintUnitStat,
} from '../../model/stat_constraint_options';

/**
 * The batch's stat constraints: one row per constraint, each a stat, an operator and a threshold,
 * e.g. Fire Res > 175. Only combinations whose final stats satisfy every row are simmed. The rows
 * live in the bulk slice; this component owns no state.
 */
export const BulkStatConstraints = () => {
	const player = usePlayer();
	const constraints = useBulkState(slice => slice.statConstraints);
	const playerClass = player.getClass();

	const replace = (idx: number, constraint: BulkStatConstraint) =>
		setBulkStatConstraints(
			player,
			constraints.map((current, i) => (i === idx ? constraint : current)),
		);

	return (
		<div data-testid="bulk-stat-constraints">
			<h6 className="mb-2">{i18n.t('bulk_tab.settings.stat_constraints.label')}</h6>
			<div className="mb-2 text-ui">{i18n.t('bulk_tab.settings.stat_constraints.tooltip')}</div>
			{!constraints.length && <div className="mb-2 text-ui text-muted">{i18n.t('bulk_tab.settings.stat_constraints.empty')}</div>}
			<div className="mb-2 grid gap-1">
				{constraints.map((constraint, idx) => {
					const unitStat = constraintUnitStat(constraint);
					const isCritReduction = unitStat.isPseudoStat() && unitStat.getPseudoStat() === PseudoStat.PseudoStatReducedCritTakenPercent;
					return (
						<div key={idx} className="flex items-center gap-1" data-testid="bulk-stat-constraints-row">
							<Select
								className="min-w-0 flex-1"
								aria-label={i18n.t('bulk_tab.settings.stat_constraints.label')}
								title={isCritReduction ? i18n.t('bulk_tab.settings.stat_constraints.crit_reduction_hint') : unitStat.getFullName(playerClass)}
								value={unitStatOptionValue(unitStat)}
								onChange={event => replace(idx, withConstraintUnitStat(constraint, unitStatFromOptionValue(event.currentTarget.value)))}>
								{SELECTABLE_STATS.map(option => (
									<option key={unitStatOptionValue(option)} value={unitStatOptionValue(option)} title={option.getFullName(playerClass)}>
										{unitStatShortLabel(option, playerClass)}
									</option>
								))}
							</Select>
							<Select
								className="w-auto"
								value={String(constraint.op)}
								onChange={event =>
									replace(idx, BulkStatConstraint.create({ ...constraint, op: Number(event.currentTarget.value) as BulkStatConstraintOp }))
								}>
								{STAT_CONSTRAINT_OPS.map(op => (
									<option key={op} value={String(op)}>
										{STAT_CONSTRAINT_OP_SYMBOLS[op]}
									</option>
								))}
							</Select>
							{/* Committed as it is typed: a value left pending until blur would commit under a
							    click on Simulate, and the combination refresh that follows disables the button. */}
							<Input
								className="w-20"
								type="number"
								step="any"
								value={String(constraint.value)}
								onChange={event => {
									const parsed = Number(event.currentTarget.value);
									replace(idx, BulkStatConstraint.create({ ...constraint, value: Number.isFinite(parsed) ? parsed : 0 }));
								}}
							/>
							<Button
								variant="link-danger"
								iconOnly
								title={i18n.t('bulk_tab.settings.stat_constraints.remove')}
								aria-label={i18n.t('bulk_tab.settings.stat_constraints.remove')}
								onClick={() =>
									setBulkStatConstraints(
										player,
										constraints.filter((_, i) => i !== idx),
									)
								}>
								<Icon name="times" />
							</Button>
						</div>
					);
				})}
			</div>
			<Button
				variant="secondary"
				size="sm"
				data-testid="bulk-stat-constraints-add"
				onClick={() => setBulkStatConstraints(player, [...constraints, newStatConstraint()])}>
				<Icon name="plus" className="mr-1" />
				{i18n.t('bulk_tab.settings.stat_constraints.add')}
			</Button>
		</div>
	);
};
