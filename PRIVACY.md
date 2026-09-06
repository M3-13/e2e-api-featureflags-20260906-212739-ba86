# Datenschutzerklärung

Dieser Dienst (`e2e-api-featureflags`) ist ein Feature-Flag-Backend ohne
Endnutzer-UI. Nachfolgend wird die Verarbeitung personenbezogener Daten gemäß
Art. 13 und Art. 30 DSGVO dokumentiert.

## Verarbeitung des `user`-Parameters

Der einzige Endpunkt, der personenbezogene Daten verarbeitet, ist
`GET /flags/{key}/evaluate?user={id}`. Der dort übergebene `user`-Parameter
wird wie folgt verarbeitet:

| Aspekt             | Angabe                                                                                     |
| ------------------ | ------------------------------------------------------------------------------------------ |
| Zweck              | Deterministische Feature-Evaluierung (stabile, wiederholbare Zuordnung eines Nutzers zu einem Rollout-Bucket) |
| Datenkategorie     | Nutzerkennung (pseudonymer Nutzer-Identifier)                                              |
| Rechtsgrundlage    | Art. 6 Abs. 1 lit. b DSGVO (Vertragserfüllung) bzw. Art. 6 Abs. 1 lit. f DSGVO (berechtigtes Interesse an reproduzierbarer Feature-Auslieferung) |
| Speicherdauer      | Keine dauerhafte Speicherung. Der Wert wird ausschließlich flüchtig im Arbeitsspeicher für die Dauer der einzelnen Anfrage gehalten und danach verworfen. |
| Empfänger          | Keine. Es erfolgt keine Weitergabe an Dritte und keine Übermittlung in Drittländer.        |

Der Dienst speichert keine Nutzer-IDs im Store und protokolliert den
`user`-Parameter nicht (die Zugriffs-Logging-Middleware erfasst ausschließlich
HTTP-Methode, Pfad und Statuscode).

## Betroffenenrechte

Betroffene Personen haben nach Maßgabe der DSGVO das Recht auf Auskunft
(Art. 15), Berichtigung (Art. 16), Löschung (Art. 17), Einschränkung der
Verarbeitung (Art. 18), Datenübertragbarkeit (Art. 20) sowie Widerspruch
(Art. 21). Da der Dienst keine personenbezogenen Daten dauerhaft speichert,
können diese Rechte in der Regel durch einfachen Verzicht auf die Übermittlung
des `user`-Parameters bzw. durch Nichtaufruf des Evaluierungs-Endpunkts erfüllt
werden. Für Anfragen wenden Sie sich an den in [SECURITY.md](./SECURITY.md)
genannten Sicherheitskontakt.

## Verantwortlicher

Der Verantwortliche im Sinne der DSGVO ist der jeweilige Betreiber der
Dienstinstanz. Der Dienst ist so konzipiert, dass er selbst möglichst wenige
personenbezogene Daten verarbeitet (Grundsatz der Datenminimierung,
Art. 5 Abs. 1 lit. c DSGVO).
