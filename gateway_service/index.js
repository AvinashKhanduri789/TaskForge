import express from "express";
import bodyParser from "body-parser";
import grpc from "@grpc/grpc-js";
import protoLoader from "@grpc/proto-loader";
import path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const app = express();
app.use(bodyParser.json());

const PROTO_PATH = path.join(__dirname, "../proto/scheduler.proto");

const packageDef = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const grpcObject = grpc.loadPackageDefinition(packageDef);
const schedulerPackage = grpcObject.scheduler;

const schedulerClient = new schedulerPackage.SchedulerService(
  "localhost:50051",
  grpc.credentials.createInsecure()
);

// ---------- Routes ----------

app.get("/health", (req, res) => {
  res.json({ status: "gateway alive" });
});

app.post("/functions", (req, res) => {
  const { name, language, code } = req.body;

  schedulerClient.RegisterFunction(
    {
      name,
      language,
      code: Buffer.from(code),
    },
    (err, response) => {
      if (err) {
        console.error(err);
        return res.status(500).json({ error: err.message });
      }

      res.json({ function_id: response.function_id });
    }
  );
});

app.post("/invoke/:functionId", (req, res) => {
  schedulerClient.TriggerExecution(
    {
      function_id: req.params.functionId,
      payload: Buffer.from(JSON.stringify(req.body)),
    },
    (err, response) => {
      if (err) {
        console.error(err);
        return res.status(500).json({ error: err.message });
      }

      res.json(response);
    }
  );
});

app.get("/jobs/:executionId", (req, res) => {
  schedulerClient.GetExecutionStatus(
    { execution_id: req.params.executionId },
    (err, response) => {
      if (err) {
        console.error(err);
        return res.status(500).json({ error: err.message });
      }

      // decode bytes -> JSON if present
      let output = null;
      if (response.output && response.output.length > 0) {
        try {
          output = JSON.parse(Buffer.from(response.output).toString());
        } catch (e) {
          output = Buffer.from(response.output).toString(); // fallback
        }
      }

      res.json({
        status: response.status,
        output,
        error: response.error,
      });
    }
  );
});


const PORT = 3000;
app.listen(PORT, () => {
  console.log(`API Gateway running on port ${PORT}`);
});
