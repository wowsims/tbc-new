import { Player } from '../../core/player.js';
import { PlayerSpecs } from '../../core/player_specs/index.js';
import { Spec } from '../../core/proto/common.js';
import { Sim } from '../../core/sim.js';
import { TypedEvent } from '../../core/typed_event.js';
import { HunterSimUI } from './sim.js';

const sim = new Sim();
const player = new Player<Spec.SpecHunter>(PlayerSpecs.Hunter, sim);
sim.raid.setPlayer(TypedEvent.nextEventID(), 0, player);

new HunterSimUI(document.body, player);
