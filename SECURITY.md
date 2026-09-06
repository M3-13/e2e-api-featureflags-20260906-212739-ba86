VERDICT: CHANGES_REQUESTED

## Security-Review: Feature-Flag-Service (Go-Backend)

**Bewertungsgrundlage:**  
Der vorliegende Code nutzt ausschließlich die Go-Standardbibliothek, speichert Flag-Daten in einem thread-sicheren In-Memory-Store und authentifiziert die `/flags`-Routen über einen Bearer-Token aus der Umgebungsvariable `FLAG_API_TOKEN`. Die Logging-Middleware protokolliert nur Methode, Pfad und Statuscode — der `user`-Parameter und andere Query-Werte erscheinen nicht im Log. Fehlerantworten sind auf `{"error":"..."}` beschränkt. Diese Punkte sind sauber umgesetzt.

Laut Scanner-Ausgabe liefen keine anwendbaren Security-Scanner. Daraus lässt sich kein Befund ableiten; die Lücke wird hier nur dokumentiert.

### Findings

#### 1. [Mittel] Fehlende Transportverschlüsselung (TLS)
- **Betroffene Stelle:** `main.go`, `http.ListenAndServe(addr, handler)`
- **Risiko:** Der Dienst überträgt den Bearer-Token im `Authorization`-Header im Klartext, sobald `BIND_ADDR` auf eine nicht-lokale Adresse (z. B. `0.0.0.0`) gesetzt wird. Der Standard `127.0.0.1` ist sicher, aber ein produktives Deployment ohne TLS-Terminierung ist angreifbar.
- **Konkreter Fix:**  
  Entweder explizit TLS in der Anwendung aktivieren (`http.ListenAndServeTLS` mit Zertifikat/Key aus Konfiguration) oder in der Deployment-Dokumentation verbindlich vorschreiben, dass der Dienst ausschließlich hinter einem TLS-terminierenden Reverse-Proxy betrieben wird. Der Standard `BIND_ADDR=127.0.0.1` sollte beibehalten werden, damit ein versehentlicher unsicherer Start nur lokal erreichbar ist.

#### 2. [Mittel] Fehlende HTTP-Server-Timeouts
- **Betroffene Stelle:** `main.go`, `http.ListenAndServe`
- **Risiko:** `http.ListenAndServe` verwendet einen `http.Server` ohne `ReadTimeout`, `WriteTimeout`, `IdleTimeout` oder `ReadHeaderTimeout`. Dadurch sind langsame Verbindungen (Slowloris) und Ressourcen-Erschöpfung möglich.
- **Konkreter Fix:**  
  Eigene `http.Server`-Instanz erzeugen und Timeouts setzen, z. B.:
  ```go
  srv := &http.Server{
      Addr:              addr,
      Handler:           handler,
      ReadHeaderTimeout: 5 * time.Second,
      ReadTimeout:       10 * time.Second,
      WriteTimeout:      10 * time.Second,
      IdleTimeout:       60 * time.Second,
  }
  if err := srv.ListenAndServe(); err != nil { log.Fatal(err) }
  ```
  Diese Werte sind großzügig genug, um den regulären Betrieb nicht zu stören.

#### 3. [Niedrig] Unbegrenzte Feldlängen bei `description` und `user`-Query
- **Betroffene Stellen:** `handlers/crud.go` (Create/Update), `handlers/evaluate.go`
- **Risiko:** Ein gültiger Bearer-Token kann 1000 Flags mit jeweils bis zu ~1 MB großen `description`-Feldern anlegen (ca. 1 GB Speicher). Ein sehr langer `user`-Query-Parameter wird in den Hash geschrieben und kann unnötig CPU/Arbeitsspeicher binden. Das ist kein direkter Angriffsvektor für Unbefugte, aber eine unnötige Ressourcenbelastung.
- **Konkreter Fix:**  
  `description` auf eine praxisnahe Maximallänge begrenzen, z. B. 4096 Zeichen, und `user` auf 256 Zeichen. Diese Limits in den Handlern vor dem Speichern bzw. Evaluieren prüfen:
  ```go
  if len(f.Description) > 4096 {
      writeError(w, http.StatusBadRequest, "description too long")
      return
  }
  if len(user) > 256 {
      writeError(w, http.StatusBadRequest, "user parameter too long")
      return
  }
  ```

#### 4. [Niedrig] JSON-Decoder ignoriert unbekannte Felder
- **Betroffene Stelle:** `handlers/crud.go`, `decodeBody`
- **Risiko:** Tippfehler in JSON-Feldern (z. B. `rollout_percentt`) werden stillschweigend ignoriert. Das kann zu Fehlkonfigurationen führen, ohne dass der Client es bemerkt.
- **Konkreter Fix:**  
  `json.Decoder` mit `DisallowUnknownFields()` verwenden:
  ```go
  dec := json.NewDecoder(r.Body)
  dec.DisallowUnknownFields()
  if err := dec.Decode(dst); err != nil { ... }
  ```
  Sofern die API-Verträge fest definiert sind, bricht dies keine legitimen Clients.

#### 5. [Niedrig] Keine Ratenbegrenzung auf der API
- **Betroffene Stelle:** `main.go`, `middleware/auth.go`
- **Risiko:** Ein Angreifer kann beliebig viele Authentifizierungsversuche mit unterschiedlichen Bearer-Tokens durchführen. Da der Token ausreichend komplex sein sollte, ist das Risiko begrenzt, aber eine Ratenbegrenzung ist eine sinnvolle Härtung.
- **Konkreter Fix:**  
  Optionale Middleware einbauen, die fehlgeschlagene Authentifizierungen pro Client-IP limitiert (z. B. Token-Bucket). Alternativ die Ratenbegrenzung im vorgelagerten Reverse-Proxy erzwingen. Die Schwelle muss so gewählt sein, dass reguläre API-Nutzung nicht beeinträchtigt wird.

#### 6. [Hinweis] Inkonsistente Key-Validierung bei GET/DELETE/Evaluate
- **Betroffene Stellen:** `handlers/crud.go` (GetFlag/DeleteFlag), `handlers/evaluate.go`
- **Risiko:** Aktuell kein direktes Sicherheitsrisiko, da der Store nur eine Map mit String-Keys verwendet. Die Key-Validierung fehlt jedoch im Gegensatz zu Create/Update. Eine einheitliche Validierung verhindert künftige Fehler, falls der Key später für andere Zwecke (z. B. Dateisystem oder externe Systeme) verwendet wird.
- **Konkreter Fix:**  
  Gemeinsame Funktion `validKey(key string) bool` nutzen und in `GetFlag`, `DeleteFlag` sowie `EvaluateFlag` vor dem Store-Zugriff prüfen. Dies ist eine rein defensive Härtung und ändert das Verhalten für gültige Keys nicht.

### Positive Befunde (keine Findings)
- **Secrets:** Kein Hardcoded Secret; `FLAG_API_TOKEN` kommt aus der Umgebung. Leerer Token führt zu Fail-Closed (401).
- **Injection/Inputs:** Request-Body-Limit 1 MB, JSON-Fehler ohne interne Details, Rollout-Validierung, Key-Pattern für Create/Update.
- **AuthN/AuthZ:** Bearer-Token-Vergleich über `crypto/subtle.ConstantTimeCompare`.
- **Datenschutz:** Logging ausschließlich Methode/Pfad/Statuscode; keine Nutzer-IDs im Store.
- **Dependencies:** Keine externen Abhängigkeiten, daher keine bekannten angreifbaren Pakete.
- **Konfiguration:** Standard-Bindung an `127.0.0.1` ist sicher.

### Fazit
Es wurden keine kritischen oder hohen Schwachstellen identifiziert. Die gefundenen Punkte sind Härtungsmaßnahmen und ein mittleres Risiko durch fehlende TLS-/Timeout-Konfiguration, das im produktiven Betrieb relevant wird. Daher: `CHANGES_REQUESTED`.