# Weryfikacja wydania

- Testy jednostkowe i integracyjne rdzenia gry: zaliczone.
- `go vet`: bez błędów.
- Windows `.exe`: poprawny plik PE32+ x86-64 z podsystemem GUI.
- Przepływ dotykowy: start, klawiatura ekranowa, odliczanie, przesunięcie i bomba sprawdzone automatycznie.
- Układ: zrzuty 3840x2160 i 2160x3840, wszystkie kluczowe obszary mieszczą się w ekranie.
- Ranking: zapis i ponowne odczytanie historii sprawdzone.
- Branding: grafika partnera została podmieniona podczas działania testu, a suma SHA-256 `.exe` pozostała identyczna.
- Ograniczenie środowiska: samego pliku PE nie uruchomiono w Windows ani Wine, ponieważ dostępne środowisko jest linuksowe, nie ma Wine i blokuje pobranie instalatora. Ten fakt nie jest zastępowany deklaracją rzeczywistego uruchomienia Windows.
