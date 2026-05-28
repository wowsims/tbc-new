import { LOCAL_STORAGE_PREFIX, REPO_RELEASES_URL } from '../../constants/other';
import type { Sim } from '../../sim';
import { isDevMode, isLocal } from '../../utils';
import { Component } from '../component';
import Toast from '../toast';
import i18n from '../../../i18n/config';

export class NoticeLocalSim extends Component {
	container: HTMLElement;
	toast: Toast | null = null;
	constructor(parent: HTMLElement, sim: Sim) {
		super(null);
		this.container = parent;

		// Synchronous opt-outs: already dismissed, or a local dev build.
		if (this.hasSeenNotice || isDevMode()) return;

		this.maybeRender(sim);
	}

	// Only recommend the native download when simulations actually run
	// in-browser via WASM. Server-hosted instances (and the downloaded app) use
	// net workers, which already compute at native speed, so the download
	// wouldn't help — regardless of whether the host is "localhost".
	private async maybeRender(sim: Sim) {
		let usingWasm: boolean;
		try {
			usingWasm = await sim.isWasm();
		} catch {
			// Workers not ready/available yet: fall back to the hostname guess so
			// we still nudge users of the public (non-local) web build.
			usingWasm = !isLocal();
		}
		if (!usingWasm || this.hasSeenNotice) return;

		this.render();

		this.toast?.element.addEventListener(
			'hide.bs.toast',
			() => {
				this.setHasSeenNotice();
			},
			{ once: true },
		);
	}

	private get settingsKey(): string {
		return `${LOCAL_STORAGE_PREFIX}_notice-local-sim.v1`;
	}

	private get hasSeenNotice() {
		return window.localStorage.getItem(this.settingsKey);
	}

	private setHasSeenNotice() {
		window.localStorage.setItem(this.settingsKey, 'true');
	}

	render() {
		this.toast = new Toast({
			additionalClasses: ['toast-notice-local-download'],
			container: this.container,
			variant: 'info',
			title: i18n.t('sim.notice_local_download.title'),
			autohide: false,
			body: (
				<div>
					<p>{i18n.t('sim.notice_local_download.message')}</p>
					<a href={REPO_RELEASES_URL} className="btn btn-outline-light" target="_blank" onclick={() => this.setHasSeenNotice()}>
						{i18n.t('sim.notice_local_download.download_button')}
					</a>
				</div>
			),
		});
	}
}
