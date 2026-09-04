# Weryfikacja wydania

- Testy jednostkowe i integracyjne rdzenia gry: zaliczone.
- `go vet`: bez błędów.
- Windows `.exe`: poprawny plik PE32+ x86-64 z podsystemem GUI.
- Przepływ dotykowy: start, klawiatura ekranowa, odliczanie, przesunięcie i bomba sprawdzone automatycznie.
- Gesty (`internal/bomber/input_test.go`): lekkie wychylenie drążka rusza postać, zmiana kierunku bez odrywania palca, histereza osi, kotwica podążająca za palcem, powrót do martwej strefy zatrzymuje marsz, bomba przy dotknięciu bez ruchu i drugim palcem, dotknięcie interfejsu działa już przy przyłożeniu palca i tylko raz, blokada zaraz po zmianie ekranu, anulowanie, wygaśnięcie i wznowienie marszu.
- Font wektorowy (`internal/bomber/font_test.go`): każdy znak rysuje piksele, `TextWidth` zgadza się z faktyczną szerokością rysowania, cache masek nie zmienia wyniku, zawijanie akapitu i dopasowanie wysokości.
- Podgląd wybuchu (`internal/bomber/game_test.go`): `BlastCells` daje dokładnie te same pola co detonacja i zatrzymuje się na ścianie i skrzyni.
- Układ nagłówka i klawiatura (`internal/bomber/render_test.go`): karty logo mieszczą się w nagłówku i nie nachodzą na siebie dla różnych proporcji grafiki i rozdzielczości; klawisze mieszczą się w swoim obszarze, a dotyk w szczelinie trafia w klawisz o najbliższym środku.
- Prymitywy rysowania (`internal/bomber/surface_test.go`): wygładzone krawędzie kół, pierścieni i zaokrąglonych prostokątów, przypadki brzegowe promienia i grubości, przycinanie poza powierzchnią.
- Tempo symulacji: `GameSpeed` skaluje krok logiki, a zegar rundy (`Elapsed`) idzie w czasie rzeczywistym (`internal/bomber/game_test.go`).
- Render 1440x810: tło i siatka planszy są buforowane; test porównuje klatkę z bufora z klatką świeżą piksel po pikselu.
- Układ: zrzuty 3840x2160 i 2160x3840, wszystkie kluczowe obszary mieszczą się w ekranie.
- Ranking: zapis i ponowne odczytanie historii sprawdzone.
- Branding: grafika partnera została podmieniona podczas działania testu, a suma SHA-256 `.exe` pozostała identyczna.
- Ograniczenie środowiska: samego pliku PE nie uruchomiono w Windows ani Wine, ponieważ dostępne środowisko jest linuksowe, nie ma Wine i blokuje pobranie instalatora. Ten fakt nie jest zastępowany deklaracją rzeczywistego uruchomienia Windows.
