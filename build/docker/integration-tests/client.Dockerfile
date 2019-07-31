FROM circleci/node:current-browsers
COPY --chown=circleci tool/chrome-with-logs /chrome-with-logs
RUN cd /chrome-with-logs; npm i
COPY --chown=circleci client/web/build /build
ENV URL=file:///build/index.html?host=berty-core&integration-tests=true&report-host=cavy-report
CMD ["node", "/chrome-with-logs"]
