VERDICT: BLOCKED

## Sicherheitsbericht

**Scanner-Abdeckung:** Für dieses `go-backend`-Projekt wurde kein anwendbarer Sicherheitsscanner ausgeführt (`(no applicable security scanners for this project type)`). Es bestehen keine Dritt-Abhängigkeiten, die ein Audit prüfen müsste. Das Fehlen von Scanner-Output ist daher kein eigenes Finding.

### 1. High — Fehlende Authentifizierung/Autorisierung auf allen API-Endpunkten
**Betroffene Stelle:** `main.go` (Routenregistrierung), `internal/handlers/crud.go` (Create/Update/Delete/Get/List), `internal/handlers/evaluate.go`

Jede Route ist ohne Authentifizierung erreichbar. Ein Angreifer mit Netzwerkzugriff auf den Port kann Feature-Flags anlegen, ändern, löschen, auslesen und Evaluierungen auslösen. Insbesondere `POST /flags`, `PUT /flags/{key}` und `DELETE /flags/{key}` erlauben unautorisierte Manipulation der Feature-Flag-Konfiguration und damit potenziell die Freischaltung oder Abschaltung von Produktfeatures. Das ist als Auth-Bypass/missing access control zu werten.

**Konkreter Fix:**
- Authentifizierungs-Middleware (z. B. API-Key oder Bearer-Token) vor alle Handler schalten, mindestens für Schreiboperationen (`POST`, `PUT`, `DELETE`) und am besten für alle Routen außer `/healthz`.
- Das Secret aus einer Umgebungsvariable lesen (nicht hardcoden), z. B. `FLAG_API_TOKEN`.
- Konfigurierbar machen, welche Endpunkte offen sind (nur für vertrauenswürdige Netze).
- Tests der Handler an die Middleware anpassen, z. B. fehlender/falscher Token → `401`.

### 2. Medium — Dienst lauscht auf allen Interfaces (`:port`) ohne TLS
**Betroffene Stelle:** `main.go`, Konfiguration

`http.ListenAndServe(":"+port, handler)` bindet standardmäßig auf alle Netzwerkschnittstellen. Wird der Dienst öffentlich oder außerhalb eines abgeschotteten Netzes betrieben, sind Flag-Daten und Verwaltungsoperationen unverschlüsselt und ohne Authentifizierung erreichbar (verschärft Finding 1). Es erfolgt keine TLS-Terminierung.

**Konkreter Fix:**
- Standardbindung auf `127.0.0.1` setzen oder eine `BIND_ADDR`-Environment-Variable einführen.
- Für externe Erreichbarkeit TLS über `http.ListenAndServeTLS` oder vorgeschalteten Reverse-Proxy erzwingen.
- Health-Check/Prometheus-Exposition bewusst und getrennt behandeln.

### 3. Low — Fehlende Begrenzung für Flag-Anzahl und Key-Größe; kein Rate-Limiting
**Betroffene Stelle:** `internal/handlers/crud.go` (Body-Limit greift nur pro Request), `internal/store/store.go`

Das 1-MB-Body-Limit begrenzt einzelne Requests, aber ein einzelner `key` kann fast 1 MB groß sein. Zudem ist die Anzahl der Flags unbegrenzt; ohne Auth kann ein Angreifer den Speicher durch viele große `POST /flags`-Requests erschöpfen (DoS).

**Konkreter Fix:**
- Maximalgröße für `key` einführen (z. B. 255 Zeichen) und prüfen.
- Optional eine maximale Anzahl von Flags konfigurieren oder Rate-Limiting vorschalten (z. B. token-bucket je Client/IP).

### 4. Low — `user`-Parameter ist frei wählbar; FNV-32a ist nicht kryptographisch
**Betroffene Stelle:** `internal/handlers/evaluate.go`, `internal/store/evaluate.go`

Der Aufrufer kann im Query-Parameter `user` beliebige Werte setzen. Da der Service nicht authentifiziert, kann ein Client die Nutzer-ID frei wählen und gezielt Hash-Buckets ausprobieren, um bei einem intermediaeren Rollout-Prozentsatz in den `true`-Bucket zu fallen. Das ist hauptsächlich eine Konsequenz der fehlenden Auth und des bewusst einfachen Hash-Verfahrens.

**Konkreter Fix:**
- Bei Einsatz für sicherheitsrelevante Features authentifizierte Nutzeridentität aus der Auth-Schicht verwenden, nicht den Client-Parameter übernehmen.
- Falls die Nutzeridentität trotzdem clientseitig übergeben werden muss, dokumentieren, dass der Service keine Sicherheitsentscheidungen tragen darf; FNV nicht durch kryptographischen Hash ersetzen, da das Rollout-Modell dadurch geändert würde.

### Positiv geprüft (kein Finding)
- **Secrets:** Keine hardcodierten Schlüssel/Passwörter/Tokens. Env-`PORT` wird ausgegeben, aber harmlos.
- **Body-Limit:** `http.MaxBytesReader` begrenzt Requests auf 1 MB; zu große Bodies werden mit definierter JSON-Fehlerantwort abgewiesen.
- **Fehlerantworten:** `writeError` liefert ausschließlich `{"error":"..."}`; keine Stacktraces, Dateipfade oder rohe Go-Fehlermeldungen.
- **Datenschutz/Logging:** Middleware protokolliert nur Methode, Pfad und Statuscode; Query-String und `user` werden nicht geloggt. Der Store speichert keine Nutzer-IDs.
- **Abhängigkeiten:** Keine externen Pakete außerhalb der Standardbibliothek; keine bekannten CVEs erkennbar.
- **Thread-Sicherheit:** Store verwendet `sync.RWMutex`; rollierende Tests und `go test -race` erscheinen abgedeckt.