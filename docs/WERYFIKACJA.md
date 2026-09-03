# Weryfikacja wydania

- Testy jednostkowe i integracyjne rdzenia gry: zaliczone.
- `go vet`: bez błędów.
- Windows `.exe`: poprawny plik PE32+ x86-64 z podsystemem GUI.
- Przepływ dotykowy: start, klawiatura ekranowa, odliczanie, przesunięcie i bomba sprawdzone automatycznie.
- Gesty (`internal/bomber/input_test.go`): ruch przed oderwaniem palca, marsz przy przytrzymaniu, zmiana kierunku, bomba drugim palcem, anulowanie, start marszu po krótkim geście.
- Prymitywy rysowania (`internal/bomber/surface_test.go`): wygładzone krawędzie kół, pierścieni i zaokrąglonych prostokątów, przypadki brzegowe promienia i grubości, przycinanie poza powierzchnią.
- Tempo symulacji: `GameSpeed` skaluje krok logiki, a zegar rundy (`Elapsed`) idzie w czasie rzeczywistym (`internal/bomber/game_test.go`).
- Render 1440x810: tło i siatka planszy są buforowane; test porównuje klatkę z bufora z klatką świeżą piksel po pikselu.
- Układ: zrzuty 3840x2160 i 2160x3840, wszystkie kluczowe obszary mieszczą się w ekranie.
- Ranking: zapis i ponowne odczytanie historii sprawdzone.
- Branding: grafika partnera została podmieniona podczas działania testu, a suma SHA-256 `.exe` pozostała identyczna.
- Ograniczenie środowiska: samego pliku PE nie uruchomiono w Windows ani Wine, ponieważ dostępne środowisko jest linuksowe, nie ma Wine i blokuje pobranie instalatora. Ten fakt nie jest zastępowany deklaracją rzeczywistego uruchomienia Windows.
