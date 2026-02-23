package ru.askorium.core.common;

import lombok.AccessLevel;
import lombok.NoArgsConstructor;
import ru.askorium.core.exception.BadUrlException;

import java.net.URI;

@NoArgsConstructor(access = AccessLevel.PRIVATE)
public class UrlUtils {

    public static String normalizeUrl(String url) throws BadUrlException {
        try {
            var uri = new URI(url);
            var normalized = new URI(
                    uri.getScheme() == null ? "https" : uri.getScheme().toLowerCase(),
                    null,
                    uri.getHost().toLowerCase(),
                    uri.getPort(),
                    uri.getPath().toLowerCase(),
                    null,
                    null
            ).normalize();
            return normalized.toString();
        } catch (Exception e) {
            throw new BadUrlException(url, e);
        }
    }
}
