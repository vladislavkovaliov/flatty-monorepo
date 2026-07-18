// @vitest-environment jsdom
import React from 'react';
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { AppComponent } from './app-component';

describe('AppComponent', () => {
  let container: HTMLElement;

  beforeEach(() => {
    container = document.createElement('div');
    document.body.appendChild(container);
  });

  afterEach(() => {
    document.body.removeChild(container);
  });

  it('creates an instance with a component', () => {
    const MockComponent = () => null;
    const instance = new AppComponent(MockComponent);
    expect(instance).toBeInstanceOf(AppComponent);
  });

  it('adds react-app-root class and renders on initialize', () => {
    const MockComponent = () => null;
    const instance = new AppComponent(MockComponent);

    instance.initialize(container, {
      env: 'development',
      featureFlags: {},
      hostType: 'react',
    });

    expect(container.classList.contains('react-app-root')).toBe(true);
  });

  it('does not throw on initialize with valid config', () => {
    const MockComponent = () => null;
    const instance = new AppComponent(MockComponent);

    expect(() => {
      instance.initialize(container, {
        env: 'production',
        featureFlags: { test: true },
        hostType: 'other',
      });
    }).not.toThrow();
  });

  it('unmounts root on destroy', () => {
    const MockComponent = () => React.createElement('div', null, 'test');
    const instance = new AppComponent(MockComponent);

    instance.initialize(container, {
      env: 'development',
      featureFlags: {},
      hostType: 'react',
    });

    expect(container.children.length).toBeGreaterThan(0);
    instance.destroy();
    expect(container.children.length).toBe(0);
  });

  it('handles destroy when not initialized', () => {
    const MockComponent = () => null;
    const instance = new AppComponent(MockComponent);

    expect(() => instance.destroy()).not.toThrow();
  });

  it('calls destroy without error twice', () => {
    const MockComponent = () => null;
    const instance = new AppComponent(MockComponent);

    instance.initialize(container, {
      env: 'development',
      featureFlags: {},
      hostType: 'react',
    });

    instance.destroy();
    expect(() => instance.destroy()).not.toThrow();
  });
});
