import { Component, OnInit, OnDestroy, HostListener } from '@angular/core';
import { ClarityIcons, helpIcon } from '@clr/angular';
import { Subscription } from 'rxjs';
import { HelperService } from '../../../services/helper/helper.service';
import { TextView } from '../../../models/content';

@Component({
  standalone: false,
  selector: 'app-helper',
  templateUrl: './helper.component.html',
  styleUrls: ['./helper.component.scss'],
})
export class HelperComponent implements OnInit, OnDestroy {
  version = '';
  commit = '';
  time = '';
  releaseInfo: TextView = {
    metadata: {
      type: 'text',
    },
    config: {
      value: '',
      isMarkdown: true,
    },
  };
  buildInfoOpen = false;
  releasesOpen = false;
  shortcutOpen = false;
  private buildInfoSubscription: Subscription;

  constructor(private helperService: HelperService) {
    ClarityIcons.addIcons(helpIcon);
  }

  ngOnInit() {
    this.buildInfoSubscription = this.helperService
      .buildVersion()
      .subscribe(version => (this.version = version));
    this.buildInfoSubscription = this.helperService
      .buildCommit()
      .subscribe(commit => (this.commit = commit));
    this.buildInfoSubscription = this.helperService
      .buildTime()
      .subscribe(time => (this.time = time));
  }

  openIssue(): void {
    window.open(
      'https://github.com/vmware-tanzu/octant/issues/new/choose',
      '_blank'
    );
  }

  getReleaseInfo(version: string) {
    const ver = version.substring(0, version.lastIndexOf('.'));
    const baseUrl =
      'https://raw.githubusercontent.com/vmware-tanzu/octant/master';
    const url =
      ver.length > 0
        ? `${baseUrl}/changelogs/CHANGELOG-${ver}.md`
        : `${baseUrl}/CHANGELOG.md`;

    fetch(url)
      .then(response => response.text())
      .then(data => {
        this.releaseInfo.config.value = data;
      });
  }

  toggleReleases(): void {
    this.getReleaseInfo(this.version);
    this.releasesOpen = !this.releasesOpen;
  }

  showDocs(): void {
    window.open('https://octant.dev/', '_blank');
  }

  ngOnDestroy(): void {
    if (this.buildInfoSubscription) {
      this.buildInfoSubscription.unsubscribe();
    }
  }

  toggleBuildInfo(): void {
    this.buildInfoOpen = !this.buildInfoOpen;
  }

  toggleShortcut(): void {
    this.shortcutOpen = !this.shortcutOpen;
  }

  @HostListener('window:keydown', ['$event'])
  keyEvent(event: KeyboardEvent) {
    if (event.ctrlKey && event.key === '/') {
      event.preventDefault();
      event.cancelBubble = true;
      this.toggleShortcut();
    }
  }
}
