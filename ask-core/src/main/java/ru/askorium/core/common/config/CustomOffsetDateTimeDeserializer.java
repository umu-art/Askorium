package ru.askorium.core.common.config;

import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;
import java.io.IOException;
import java.time.OffsetDateTime;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeParseException;
import java.util.List;
import org.springframework.stereotype.Component;

@Component
public class CustomOffsetDateTimeDeserializer extends JsonDeserializer<OffsetDateTime> {

    private final List<DateTimeFormatter> formatters = List.of(
            DateTimeFormatter.ISO_OFFSET_DATE_TIME,
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss.SSSSSSXXX"),
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss.SSSXXX"),
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ssXXX")
    );

    @Override
    public OffsetDateTime deserialize(JsonParser jsonParser, DeserializationContext deserializationContext) throws IOException {
        var text = jsonParser.getText();
        for (var formatter : formatters) {
            try {
                return OffsetDateTime.parse(text, formatter);
            } catch (DateTimeParseException ignored) {
            }
        }
        throw new IOException("Unparseable OffsetDateTime: " + text);
    }
}
