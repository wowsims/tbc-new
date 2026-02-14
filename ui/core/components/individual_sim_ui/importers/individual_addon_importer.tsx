import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../../individual_sim_ui';
import { Class, EquipmentSpec, Profession, Race, Spec } from '../../../proto/common';
import { Database } from '../../../proto_utils/database';
import { nameToClass, nameToProfession, nameToRace } from '../../../proto_utils/names';
import Toast from '../../toast';
import { IndividualImporter } from './individual_importer';
import i18n from '../../../../i18n/config';

export class IndividualAddonImporter<SpecType extends Spec> extends IndividualImporter<SpecType> {
	static WSE_VERSION = getWSEVersion();
	constructor(parent: HTMLElement, simUI: IndividualSimUI<SpecType>) {
		super(parent, simUI, { title: i18n.t('import.addon.title'), allowFileUpload: true });

		const warningRef = ref<HTMLDivElement>();
		this.descriptionElem.appendChild(
			<div>
				<p>
					{i18n.t('import.addon.description')}{' '}
					<a href="https://www.curseforge.com/wow/addons/wowsimsexporter" target="_blank">
						{i18n.t('import.addon.addon_link')}
					</a>
					.
				</p>
				<p>{i18n.t('import.addon.feature_description')}</p>
				<p>{i18n.t('import.addon.instructions')}</p>
				<div ref={warningRef} />
			</div>
		);
	}

	async onImport(data: string) {
		let importJson: any | null;
		try {
			importJson = JSON.parse(data);
		} catch {
			throw new Error('Please use a valid Addon export.');
		}

		let addonVersion = await IndividualAddonImporter.WSE_VERSION;
		if (addonVersion && ((importJson['version'] as string) || '') != addonVersion) {
			new Toast({ variant: 'warning', body: `Addon is not up to date. Addon version : '${importJson['version']}', Latest version : '${addonVersion}'` });
		}

		// Parse all the settings.
		const charClass = nameToClass((importJson['class'] as string) || '');
		if (charClass == Class.ClassUnknown) {
			throw new Error('Could not parse Class!');
		}

		const race = nameToRace((importJson['race'] as string) || '');
		if (race == Race.RaceUnknown) {
			throw new Error('Could not parse Race!');
		}

		const professions = (importJson['professions'] as Array<{ name: string; level: number }>).map(profData => nameToProfession(profData.name));
		professions.forEach((prof, i) => {
			if (prof == Profession.ProfessionUnknown) {
				throw new Error(`Could not parse profession '${importJson['professions'][i]}'`);
			}
		});

		const talentsStr = (importJson['talents'] as string) || '';

		const db = await Database.get();

		const gearJson = importJson['gear'];
		gearJson.items = (gearJson.items as Array<any>).filter(item => item != null);
		delete gearJson.version;

		(gearJson.items as Array<any>).forEach(item => {
			if (item.gems) {
				item.gems = (item.gems as Array<any>).map(gem => gem || 0);
			}
		});
		const equipmentSpec = EquipmentSpec.fromJson(gearJson);

		this.finishIndividualImport(this.simUI, {
			charClass,
			race,
			equipmentSpec,
			talentsStr,
			professions,
		});
	}
}

function getWSEVersion(): Promise<string|null> {
	return fetch('https://api.github.com/repos/wowsims/exporter/releases/latest')
		.then(resp => {
			return resp.json().then(json => {
				return json.tag_name as string;
			})
		})
		.catch(_ => {
			return null;
		})
}
