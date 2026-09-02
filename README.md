# BOMBER RUSH

Gotowa gra kioskowa dla Windows x86-64. Domyślnie uruchamia się bez ramek, na pełnym ekranie i ponad innymi oknami. Cała obsługa odbywa się dotykiem:

- przesunięcie palcem w czterech kierunkach - ruch; ruch zaczyna się już w trakcie gestu, a przytrzymany palec prowadzi postać dalej,
- zmiana kierunku - przeciągnięcie palca w nową stronę bez odrywania,
- dotknięcie planszy albo drugi palec podczas marszu - ustawienie bomby (na start dwie bomby naraz),
- pierwsza runda - krótka podpowiedź gestów,
- wynik rundy pozostaje na ekranie do wybrania rankingu albo ponownej gry,
- nick, ranking i panel historii - duże przyciski ekranowe.

Gra używa krótkich dźwięków systemowych Windows dla najważniejszych zdarzeń. Jeśli katalog `data` nie jest zapisywalny, pokazuje ostrzeżenie zamiast ukrywać problem.

## Uruchomienie

Zachowaj razem plik `BomberRush.exe` oraz katalogi `assets` i `data`, a następnie uruchom `BomberRush.exe`. Wyniki są dopisywane lokalnie do `data/scores.jsonl`.

Tryb diagnostyczny w oknie:

```bat
BomberRush.exe --windowed --width 1600 --height 900
```

## Podmiana materiałów partnera bez rekompilacji

1. Zastąp odpowiednie pliki PNG w katalogu `assets` albo wpisz ich nowe nazwy w `assets/branding.json`.
2. Zachowaj pliki wewnątrz katalogu `assets`.
3. Gra wykryje zmianę podczas działania, zwykle w ciągu sekundy.

Obsługiwane materiały:

| Plik domyślny | Zastosowanie | Zalecany format |
|---|---|---|
| `partner-logo.png` | logo partnera w nagłówku | PNG z przezroczystością, 1200x360 |
| `start-background.png` | tło ekranu startowego | PNG, 1920x1080 lub większe |
| `summary-ad.png` | reklama na podsumowaniu | PNG, 1200x700 |
| `qr-placeholder.png` | kod QR lub kupon | PNG, kwadrat min. 800x800 |
| `product-item.png` | przedmiot zbierany na planszy | PNG z przezroczystością, min. 256x256 |

W `branding.json` można zmienić nazwę wydarzenia, teksty, kolory, identyfikator rankingu, PIN administratora oraz czas rundy: 60, 75 albo 90 sekund. Zmiana `event_id` tworzy osobny ranking wydarzenia bez usuwania historii.

Logo nanoVo jest stałym, wbudowanym zasobem i nie podlega konfiguracji partnera. Oryginalny załącznik był plikiem SVG; jego wiernie zrasteryzowana wersja PNG znajduje się w `supplied-logo`.

## Budowa Windows `.exe`

Wymagany jest Go 1.23 lub nowszy. Projekt nie używa zewnętrznych bibliotek Go ani CGO.

Windows:

```bat
build-windows.bat
```

Linux lub macOS:

```bash
./scripts/build-windows.sh
```

Wynikiem jest `dist/BomberRush.exe`.

## CI

Push na `beta` uruchamia `.github/workflows/ci.yml` na runnerach self-hosted: testy, `go vet`, build `dist/BomberRush.exe` jako artefakt. Po zielonym buildzie workflow fast-forwarduje `main`, tworzy tag `vX.Y.Z` z pliku `VERSION` i publikuje release z `.exe`. Numer w `VERSION` podnosi się przed wydaniem.

## Weryfikacja deweloperska

Na Linuxie:

```bash
./scripts/verify-linux.sh
```

Skrypt uruchamia testy, `go vet`, buduje Windows PE i generuje raport oraz zrzuty 4K w katalogu `evidence`.
