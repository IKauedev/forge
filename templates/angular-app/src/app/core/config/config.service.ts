import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';
import { AppConfig } from './app-config';

/** Configuração de runtime vinda de /config.json (servido sem cache, ao lado do index.html). */
@Injectable({ providedIn: 'root' })
export class ConfigService {
  private readonly http = inject(HttpClient);
  private readonly state = signal<AppConfig | null>(null);

  readonly config = computed(() => {
    const value = this.state();
    if (!value) {
      throw new Error('Configuração ainda não carregada (config.json).');
    }
    return value;
  });

  async load(): Promise<void> {
    this.state.set(await firstValueFrom(this.http.get<AppConfig>('config.json')));
  }
}
