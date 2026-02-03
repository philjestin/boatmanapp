import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MainPanel } from './MainPanel';

describe('MainPanel', () => {
  it('renders empty state when isEmpty is true', () => {
    const mockOnNewSession = vi.fn();
    const mockOnOpenProject = vi.fn();

    render(
      <MainPanel
        isEmpty={true}
        onNewSession={mockOnNewSession}
        onOpenProject={mockOnOpenProject}
      />
    );

    expect(screen.getByText('Welcome to Boatman')).toBeInTheDocument();
    expect(screen.getByText(/Start a new session or open a project/)).toBeInTheDocument();
    expect(screen.getByText('Open Project')).toBeInTheDocument();
    expect(screen.getByText('New Session')).toBeInTheDocument();
  });

  it('calls onOpenProject when Open Project button is clicked', () => {
    const mockOnNewSession = vi.fn();
    const mockOnOpenProject = vi.fn();

    render(
      <MainPanel
        isEmpty={true}
        onNewSession={mockOnNewSession}
        onOpenProject={mockOnOpenProject}
      />
    );

    fireEvent.click(screen.getByText('Open Project'));
    expect(mockOnOpenProject).toHaveBeenCalledTimes(1);
  });

  it('calls onNewSession when New Session button is clicked', () => {
    const mockOnNewSession = vi.fn();
    const mockOnOpenProject = vi.fn();

    render(
      <MainPanel
        isEmpty={true}
        onNewSession={mockOnNewSession}
        onOpenProject={mockOnOpenProject}
      />
    );

    fireEvent.click(screen.getByText('New Session'));
    expect(mockOnNewSession).toHaveBeenCalledTimes(1);
  });

  it('renders children when isEmpty is false', () => {
    render(
      <MainPanel isEmpty={false}>
        <div data-testid="child-content">Test Content</div>
      </MainPanel>
    );

    expect(screen.getByTestId('child-content')).toBeInTheDocument();
    expect(screen.getByText('Test Content')).toBeInTheDocument();
  });

  it('renders as a main element with correct structure', () => {
    const { container } = render(
      <MainPanel isEmpty={false}>
        <div>Content</div>
      </MainPanel>
    );

    const main = container.querySelector('main');
    expect(main).toBeInTheDocument();
    expect(main).toHaveClass('flex-1', 'flex', 'flex-col', 'overflow-hidden');
  });
});
