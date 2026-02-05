/**
 * Search Bar Component
 * VE-703: Marketplace search input with suggestions
 */
import * as React from 'react';
import { useState, useCallback, useRef, useEffect } from 'react';

export interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  onSearch: (value: string) => void;
  placeholder?: string;
  suggestions?: string[];
  isLoading?: boolean;
  className?: string;
}

export function SearchBar({
  value,
  onChange,
  onSearch,
  placeholder = 'Search offerings...',
  suggestions = [],
  isLoading = false,
  className = '',
}: SearchBarProps): JSX.Element {
  const [isFocused, setIsFocused] = useState(false);
  const [selectedSuggestion, setSelectedSuggestion] = useState(-1);
  const inputRef = useRef<HTMLInputElement>(null);
  const suggestionsRef = useRef<HTMLUListElement>(null);

  const showSuggestions = isFocused && suggestions.length > 0 && value.length > 0;

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      onSearch(value);
      setIsFocused(false);
      inputRef.current?.blur();
    },
    [value, onSearch]
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!showSuggestions) return;

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedSuggestion((prev) =>
            prev < suggestions.length - 1 ? prev + 1 : prev
          );
          break;
        case 'ArrowUp':
          e.preventDefault();
          setSelectedSuggestion((prev) => (prev > 0 ? prev - 1 : -1));
          break;
        case 'Enter':
          if (selectedSuggestion >= 0) {
            e.preventDefault();
            onChange(suggestions[selectedSuggestion]);
            onSearch(suggestions[selectedSuggestion]);
            setIsFocused(false);
          }
          break;
        case 'Escape':
          setIsFocused(false);
          inputRef.current?.blur();
          break;
      }
    },
    [showSuggestions, suggestions, selectedSuggestion, onChange, onSearch]
  );

  const handleSuggestionClick = useCallback(
    (suggestion: string) => {
      onChange(suggestion);
      onSearch(suggestion);
      setIsFocused(false);
    },
    [onChange, onSearch]
  );

  // Reset selection when suggestions change
  useEffect(() => {
    setSelectedSuggestion(-1);
  }, [suggestions]);

  const searchId = 'marketplace-search';
  const suggestionsId = `${searchId}-suggestions`;

  return (
    <form
      className={`search-bar ${className}`}
      onSubmit={handleSubmit}
      role="search"
      aria-label="Search marketplace offerings"
    >
      <div className="search-bar__container">
        <span className="search-bar__icon" aria-hidden="true">
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35" />
          </svg>
        </span>

        <input
          ref={inputRef}
          id={searchId}
          type="search"
          className="search-bar__input"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setTimeout(() => setIsFocused(false), 150)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          autoComplete="off"
          aria-label="Search"
          aria-autocomplete="list"
          aria-controls={showSuggestions ? suggestionsId : undefined}
          aria-expanded={showSuggestions}
          aria-activedescendant={
            selectedSuggestion >= 0
              ? `${suggestionsId}-${selectedSuggestion}`
              : undefined
          }
        />

        {isLoading && (
          <span className="search-bar__loading" aria-label="Searching">
            <svg
              className="search-bar__spinner"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <circle cx="12" cy="12" r="10" opacity="0.25" />
              <path d="M12 2a10 10 0 0 1 10 10" />
            </svg>
          </span>
        )}

        {value && !isLoading && (
          <button
            type="button"
            className="search-bar__clear"
            onClick={() => {
              onChange('');
              inputRef.current?.focus();
            }}
            aria-label="Clear search"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
        )}

        <button type="submit" className="search-bar__submit" aria-label="Search">
          Search
        </button>
      </div>

      {showSuggestions && (
        <ul
          ref={suggestionsRef}
          id={suggestionsId}
          className="search-bar__suggestions"
          role="listbox"
          aria-label="Search suggestions"
        >
          {suggestions.map((suggestion, index) => (
            <li
              key={suggestion}
              id={`${suggestionsId}-${index}`}
              className={`search-bar__suggestion ${
                index === selectedSuggestion ? 'search-bar__suggestion--selected' : ''
              }`}
              role="option"
              aria-selected={index === selectedSuggestion}
              onClick={() => handleSuggestionClick(suggestion)}
            >
              <span className="search-bar__suggestion-icon" aria-hidden="true">
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <circle cx="11" cy="11" r="8" />
                  <path d="M21 21l-4.35-4.35" />
                </svg>
              </span>
              {suggestion}
            </li>
          ))}
        </ul>
      )}

      <style>{searchBarStyles}</style>
    </form>
  );
}

const searchBarStyles = `
  .search-bar {
    position: relative;
    width: 100%;
    max-width: 600px;
  }

  .search-bar__container {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  .search-bar__container:focus-within {
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .search-bar__icon {
    color: #9ca3af;
    flex-shrink: 0;
    display: flex;
  }

  .search-bar__input {
    flex: 1;
    border: none;
    outline: none;
    font-size: 1rem;
    color: #111827;
    background: transparent;
    min-width: 0;
  }

  .search-bar__input::placeholder {
    color: #9ca3af;
  }

  .search-bar__input::-webkit-search-cancel-button {
    display: none;
  }

  .search-bar__loading {
    color: #3b82f6;
    display: flex;
  }

  .search-bar__spinner {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .search-bar__clear {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 4px;
    background: transparent;
    border: none;
    border-radius: 4px;
    color: #6b7280;
    cursor: pointer;
    transition: color 0.2s, background-color 0.2s;
  }

  .search-bar__clear:hover {
    color: #111827;
    background: #f3f4f6;
  }

  .search-bar__submit {
    padding: 8px 16px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
    flex-shrink: 0;
  }

  .search-bar__submit:hover {
    background: #2563eb;
  }

  .search-bar__submit:focus-visible {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  .search-bar__suggestions {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    right: 0;
    margin: 0;
    padding: 8px 0;
    list-style: none;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    z-index: 50;
    max-height: 300px;
    overflow-y: auto;
  }

  .search-bar__suggestion {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 16px;
    cursor: pointer;
    transition: background-color 0.1s;
  }

  .search-bar__suggestion:hover,
  .search-bar__suggestion--selected {
    background: #f3f4f6;
  }

  .search-bar__suggestion-icon {
    color: #9ca3af;
    display: flex;
    flex-shrink: 0;
  }

  @media (prefers-reduced-motion: reduce) {
    .search-bar__container,
    .search-bar__clear,
    .search-bar__submit,
    .search-bar__suggestion {
      transition: none;
    }
    .search-bar__spinner {
      animation: none;
    }
  }

  @media (max-width: 640px) {
    .search-bar__submit {
      padding: 8px 12px;
      font-size: 0.8125rem;
    }
  }
`;
