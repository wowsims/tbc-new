import * as InputHelpers from '../core/components/input_helpers';
import { Player } from '../core/player';
import { UnitReference, UnitReference_Type as UnitType } from '../core/proto/common';
import { ActionId } from '../core/proto_utils/action_id';
import { DruidSpecs } from '../core/proto_utils/utils';
import { EventID } from '../core/typed_event';

// Configuration for class-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.
