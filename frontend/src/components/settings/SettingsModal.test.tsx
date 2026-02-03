import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SettingsModal } from './SettingsModal';
import type { UserPreferences } from '../../types';

describe('SettingsModal', () => {
  const defaultPreferences: UserPreferences = {
    apiKey: '',
    approvalMode: 'suggest',
    defaultModel: 'claude-sonnet-4-20250514',
    theme: 'dark',
    notificationsEnabled: true,
    mcpServers: [],
    onboardingCompleted: true,
  };

  const mockOnClose = vi.fn();
  const mockOnSave = vi.fn();

  beforeEach(() => {
    mockOnClose.mockClear();
    mockOnSave.mockClear();
  });

  it('renders nothing when isOpen is false', () => {
    const { container } = render(
      <SettingsModal
        isOpen={false}
        onClose={mockOnClose}
        preferences={defaultPreferences}
        onSave={mockOnSave}
      />
    );
    expect(container.firstChild).toBeNull();
  });

  it('renders settings modal when isOpen is true', () => {
    render(
      <SettingsModal
        isOpen={true}
        onClose={mockOnClose}
        preferences={defaultPreferences}
        onSave={mockOnSave}
      />
    );

    expect(screen.getByText('Settings')).toBeInTheDocument();
    expect(screen.getByText('General')).toBeInTheDocument();
    expect(screen.getByText('Approval')).toBeInTheDocument();
    expect(screen.getByText('MCP Servers')).toBeInTheDocument();
    expect(screen.getByText('About')).toBeInTheDocument();
  });

  it('calls onClose when Cancel is clicked', () => {
    render(
      <SettingsModal
        isOpen={true}
        onClose={mockOnClose}
        preferences={defaultPreferences}
        onSave={mockOnSave}
      />
    );

    fireEvent.click(screen.getByText('Cancel'));
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('calls onSave and onClose when Save Changes is clicked', () => {
    render(
      <SettingsModal
        isOpen={true}
        onClose={mockOnClose}
        preferences={defaultPreferences}
        onSave={mockOnSave}
      />
    );

    fireEvent.click(screen.getByText('Save Changes'));
    expect(mockOnSave).toHaveBeenCalledTimes(1);
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('displays general settings tab by default', () => {
    render(
      <SettingsModal
        isOpen={true}
        onClose={mockOnClose}
        preferences={defaultPreferences}
        onSave={mockOnSave}
      />
    );

    expect(screen.getByText('API Key')).toBeInTheDocument();
    expect(screen.getByText('Appearance')).toBeInTheDocument();
    expect(screen.getByText('Notifications')).toBeInTheDocument();
    expect(screen.getByText('Default Model')).toBeInTheDocument();
  });

  it('switches to Approval tab when clicked', () => {
    render(
      <SettingsModal
        isOpen={true}
        onClose={mockOnClose}
        preferences={defaultPreferences}
        onSave={mockOnSave}
      />
    );

    fireEvent.click(screen.getByText('Approval'));

    expect(screen.getByText('Approval Mode')).toBeInTheDocument();
    expect(screen.getByText('Suggest Mode')).toBeInTheDocument();
    expect(screen.getByText('Auto-Edit Mode')).toBeInTheDocument();
    expect(screen.getByText('Full Auto Mode')).toBeInTheDocument();
  });

  it('switches to About tab when clicked', () => {
    render(
      <SettingsModal
        isOpen={true}
        onClose={mockOnClose}
        preferences={defaultPreferences}
        onSave={mockOnSave}
      />
    );

    fireEvent.click(screen.getByText('About'));

    expect(screen.getByText('Boatman')).toBeInTheDocument();
    expect(screen.getByText('Version 0.1.0')).toBeInTheDocument();
  });

  it('shows warning when full-auto mode is selected', () => {
    render(
      <SettingsModal
        isOpen={true}
        onClose={mockOnClose}
        preferences={{ ...defaultPreferences, approvalMode: 'full-auto' }}
        onSave={mockOnSave}
      />
    );

    fireEvent.click(screen.getByText('Approval'));

    expect(screen.getByText(/Warning: Full auto mode gives Claude complete control/)).toBeInTheDocument();
  });

  it('does not show warning for suggest mode', () => {
    render(
      <SettingsModal
        isOpen={true}
        onClose={mockOnClose}
        preferences={{ ...defaultPreferences, approvalMode: 'suggest' }}
        onSave={mockOnSave}
      />
    );

    fireEvent.click(screen.getByText('Approval'));

    expect(screen.queryByText(/Warning: Full auto mode gives Claude complete control/)).not.toBeInTheDocument();
  });
});
