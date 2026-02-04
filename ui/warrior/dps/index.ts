import { Player } from '../../core/player';
import { PlayerSpecs } from '../../core/player_specs';
import { Spec } from '../../core/proto/common';
import { Sim } from '../../core/sim';
import { TypedEvent } from '../../core/typed_event';
import { DpsWarriorSimUI } from './sim';

const sim = new Sim();
const player = new Player<Spec.SpecDpsWarrior>(PlayerSpecs.DpsWarrior, sim);
sim.raid.setPlayer(TypedEvent.nextEventID(), 0, player);

new DpsWarriorSimUI(document.body, player);
