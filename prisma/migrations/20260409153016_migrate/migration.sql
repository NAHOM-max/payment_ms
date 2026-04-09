/*
  Warnings:

  - Changed the type of `status` on the `payment` table. No cast exists, the column would be dropped and recreated, which cannot be done if there is data, since the column is required.

*/
-- CreateEnum
CREATE TYPE "PaymentStatus" AS ENUM ('INITIATED', 'SUCCESSFUL', 'FAILED', 'CANCELED');

-- AlterTable
ALTER TABLE "payment" DROP COLUMN "status",
ADD COLUMN     "status" "PaymentStatus" NOT NULL;

-- CreateIndex
CREATE INDEX "payment_workflowId_idx" ON "payment"("workflowId");
