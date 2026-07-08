import { Component, ElementRef, OnInit, OnDestroy, inject } from '@angular/core';
import { loadBundle } from './loader-utils';
import type { IAppComponent } from '@flatty-budget/shared';

@Component({
  selector: 'app-external',
  standalone: true,
  template: `<div #host></div>`,
})
export class ExternalComponent implements OnInit, OnDestroy {
  private readonly hostRef = inject(ElementRef<HTMLDivElement>);
  private app: IAppComponent | undefined;
  private readonly host: HTMLDivElement | undefined;

  ngOnInit(): void {
    const host = this.hostRef.nativeElement.querySelector('div');
    if (!host) return;

    loadBundle('ext-apps', 'settings', '/external-settings')
      .then((module) => {
        this.app = module;
        this.app!.initialize(host, {});
      })
      .catch(console.error);
  }

  ngOnDestroy(): void {
    this.app?.destroy();
  }
}
