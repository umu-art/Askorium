FROM python:3.14-slim
LABEL authors="vikazeni"

WORKDIR /application

COPY /ask-encoder/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY /api/build/python-encoder-api/dist/python_encoder_api-*-py3-none-any.whl .
RUN pip install --no-cache-dir ./*.whl

COPY ask-encoder .

RUN python -m app --fetch

ENTRYPOINT ["python", "-m", "app"]
