/*
 * Copyright (c) 2020 the Octant contributors. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Injectable } from '@angular/core';
import { ipcRenderer, webFrame, webUtils } from 'electron';
import * as childProcess from 'child_process';
import * as fs from 'fs';
import { PreferencesService } from '../preferences/preferences.service';

@Injectable({
  providedIn: 'root',
})
export class ElectronService {
  ipcRenderer: typeof ipcRenderer;
  webFrame: typeof webFrame;
  webUtils: typeof webUtils;
  childProcess: typeof childProcess;
  fs: typeof fs;

  public portNumber: number;
  constructor(private preferencesService: PreferencesService) {
    if (this.isElectron()) {
      this.ipcRenderer = window.require('electron').ipcRenderer;
      this.webFrame = window.require('electron').webFrame;
      this.webUtils = window.require('electron').webUtils;
      this.childProcess = window.require('child_process');
      this.fs = window.require('fs');

      this.ipcRenderer.once('port-message', (event, message) => {
        this.portNumber = message;
      });

      this.ipcRenderer.on('openPreferences', () => {
        if (!this.preferencesService.preferencesOpened.value) {
          this.preferencesService.preferencesOpened.next(true);
        }
      });
    }
  }

  /**
   * Returns true if electron is detected
   */
  isElectron(): boolean {
    return this.preferencesService.isElectron();
  }

  /**
   * Returns the random port number from electron main process
   */
  port(): number {
    return this.portNumber;
  }

  /**
   * Returns the platform.
   *   * Returns linux, darwin, or win32 for those platforms
   *   * Returns unknown if the platform is not linux, darwin, or win32
   *   * Returns a blank string is electron is not detected
   *
   */
  platform(): string {
    if (!this.isElectron()) {
      return '';
    }

    switch (process.platform) {
      case 'linux':
      case 'darwin':
      case 'win32':
        return process.platform;
      default:
        return 'unknown';
    }
  }

  /**
   * Returns the OS path for a File picked in the renderer. Electron removed the
   * `File.path` property in v32; `webUtils.getPathForFile` replaces it.
   */
  getPathForFile(file: File): string {
    if (!this.isElectron()) {
      return undefined;
    }
    if (this.webUtils && this.webUtils.getPathForFile) {
      return this.webUtils.getPathForFile(file);
    }
    // Fallback for older Electron versions that still expose File.path.
    return (file as File & { path?: string }).path;
  }
}
